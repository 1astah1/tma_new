package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type OrderService struct {
	db         *sqlx.DB
	orderRepo  *repository.OrderRepo
	productRepo *repository.ProductRepo
	keyRepo    *repository.KeyRepo
	accountRepo *repository.AccountRepo
	userRepo   *repository.UserRepo
	encSvc     *EncryptionService
	notifSvc   *NotificationService
	auditSvc   *AuditService
}

func NewOrderService(
	db *sqlx.DB,
	orderRepo *repository.OrderRepo,
	productRepo *repository.ProductRepo,
	keyRepo *repository.KeyRepo,
	accountRepo *repository.AccountRepo,
	userRepo *repository.UserRepo,
	encSvc *EncryptionService,
	notifSvc *NotificationService,
	auditSvc *AuditService,
) *OrderService {
	return &OrderService{
		db: db, orderRepo: orderRepo, productRepo: productRepo,
		keyRepo: keyRepo, accountRepo: accountRepo, userRepo: userRepo,
		encSvc: encSvc, notifSvc: notifSvc, auditSvc: auditSvc,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, productID uuid.UUID, deliveryMethod domain.DeliveryMethod) (*domain.Order, error) {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if product.Status != domain.ProductStatusActive {
		return nil, domain.ErrInvalidInput
	}

	valid := false
	for _, dm := range product.DeliveryMethods {
		if dm == string(deliveryMethod) {
			valid = true
			break
		}
	}
	if !valid {
		return nil, domain.ErrInvalidInput
	}

	order := &domain.Order{
		UserID:         userID,
		ProductID:      productID,
		DeliveryMethod: deliveryMethod,
		Status:         domain.OrderStatusNew,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Transition to WAITING_PAYMENT
	s.changeStatus(ctx, order, domain.OrderStatusWaitingPayment, nil, domain.ChangedBySystem, "Order created")

	go func() {
		s.notifSvc.SendOrderStatusUpdate(context.Background(), order)
	}()

	return order, nil
}

func (s *OrderService) ConfirmPayment(ctx context.Context, orderID, adminID uuid.UUID, paymentMethod string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusPaymentVerification {
		return domain.ErrOrderStatusInvalid
	}

	order.PaymentMethod = &paymentMethod
	order.PaymentVerifiedBy = &adminID
	s.orderRepo.Update(ctx, order)

	return s.changeStatus(ctx, order, domain.OrderStatusPaid, &adminID, domain.ChangedByAdmin, "Payment confirmed")
}

func (s *OrderService) IssueKey(ctx context.Context, orderID, adminID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusPaid || order.DeliveryMethod != domain.DeliveryMethodKey {
		return domain.ErrOrderStatusInvalid
	}

	key, err := s.keyRepo.AssignAvailableKey(ctx, order.ProductID, orderID)
	if err != nil {
		return err
	}

	order.KeyID = &key.ID
	s.orderRepo.Update(ctx, order)

	if err := s.changeStatus(ctx, order, domain.OrderStatusKeyIssued, &adminID, domain.ChangedByAdmin, "Key issued"); err != nil {
		return err
	}

	return s.changeStatus(ctx, order, domain.OrderStatusCompleted, &adminID, domain.ChangedBySystem, "Order completed")
}

func (s *OrderService) AssignActivation(ctx context.Context, orderID, adminID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusWaitingActivation {
		return domain.ErrOrderStatusInvalid
	}

	order.AssignedAdminID = &adminID
	s.orderRepo.Update(ctx, order)

	return s.changeStatus(ctx, order, domain.OrderStatusAwaitingCredentials, &adminID, domain.ChangedByAdmin, "Task assigned")
}

func (s *OrderService) ReceiveCredentials(ctx context.Context, orderID, userID uuid.UUID, platform domain.Platform, login, password string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusAwaitingCredentials {
		return domain.ErrOrderStatusInvalid
	}

	encLogin, err := s.encSvc.Encrypt([]byte(login))
	if err != nil {
		return err
	}
	encPassword, err := s.encSvc.Encrypt([]byte(password))
	if err != nil {
		return err
	}

	account := &domain.UserAccount{
		UserID:   userID,
		OrderID:  orderID,
		Platform: platform,
		Login:    encLogin,
		Password: encPassword,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return err
	}

	return s.changeStatus(ctx, order, domain.OrderStatusCredentialsReceived, nil, domain.ChangedByUser, "Credentials received")
}

func (s *OrderService) Request2FA(ctx context.Context, orderID, adminID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusCredentialsReceived {
		return domain.ErrOrderStatusInvalid
	}

	return s.changeStatus(ctx, order, domain.OrderStatusAwaiting2FA, &adminID, domain.ChangedByAdmin, "2FA requested")
}

func (s *OrderService) Receive2FA(ctx context.Context, orderID uuid.UUID, code string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusAwaiting2FA {
		return domain.ErrOrderStatusInvalid
	}

	account, err := s.accountRepo.GetByOrderID(ctx, orderID)
	if err == nil {
		s.accountRepo.Update2FACode(ctx, account.ID, &code)
	}

	return s.changeStatus(ctx, order, domain.OrderStatusActivating, nil, domain.ChangedByUser, "2FA code received")
}

func (s *OrderService) CompleteActivation(ctx context.Context, orderID, adminID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusActivating {
		return domain.ErrOrderStatusInvalid
	}

	if err := s.changeStatus(ctx, order, domain.OrderStatusActivated, &adminID, domain.ChangedByAdmin, "Activation completed"); err != nil {
		return err
	}

	return s.changeStatus(ctx, order, domain.OrderStatusCompleted, &adminID, domain.ChangedBySystem, "Order completed")
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID, adminID uuid.UUID, reason string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	order.CancelledReason = &reason
	s.orderRepo.Update(ctx, order)

	return s.changeStatus(ctx, order, domain.OrderStatusCancelled, &adminID, domain.ChangedByAdmin, reason)
}

func (s *OrderService) RequestRefund(ctx context.Context, orderID, adminID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if !canRefund(order.Status) {
		return domain.ErrOrderStatusInvalid
	}

	return s.changeStatus(ctx, order, domain.OrderStatusRefundRequested, &adminID, domain.ChangedByAdmin, "Refund requested")
}

func (s *OrderService) ProcessRefund(ctx context.Context, orderID, adminID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusRefundRequested {
		return domain.ErrOrderStatusInvalid
	}

	return s.changeStatus(ctx, order, domain.OrderStatusRefunded, &adminID, domain.ChangedByAdmin, "Refund processed")
}

func canRefund(status domain.OrderStatus) bool {
	switch status {
	case domain.OrderStatusPaid, domain.OrderStatusWaitingActivation,
		domain.OrderStatusAwaitingCredentials, domain.OrderStatusCredentialsReceived,
		domain.OrderStatusAwaiting2FA, domain.OrderStatusActivating,
		domain.OrderStatusActivated, domain.OrderStatusKeyIssued:
		return true
	}
	return false
}

func (s *OrderService) changeStatus(ctx context.Context, order *domain.Order, newStatus domain.OrderStatus, changedByID *uuid.UUID, changedByType domain.ChangedByType, comment string) error {
	if !domain.IsValidTransition(order.Status, newStatus) {
		return domain.ErrOrderStatusInvalid
	}

	oldStatus := order.Status
	order.Status = newStatus

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, newStatus); err != nil {
		return err
	}

	history := &domain.OrderHistory{
		OrderID:       order.ID,
		OldStatus:     &oldStatus,
		NewStatus:     newStatus,
		ChangedByID:   changedByID,
		ChangedByType: changedByType,
		Comment:       &comment,
	}
	if err := s.orderRepo.AddHistory(ctx, history); err != nil {
		return err
	}

	if changedByID != nil && changedByType == domain.ChangedByAdmin {
		s.auditSvc.Log(ctx, *changedByID, "order_status_change", "order", order.ID,
			map[string]interface{}{
				"old_status": oldStatus,
				"new_status": newStatus,
				"comment":    comment,
			})
	}

	// Notify user
	go func() {
		updatedOrder, err := s.orderRepo.GetByIDWithJoins(context.Background(), order.ID)
		if err == nil {
			s.notifSvc.SendOrderStatusUpdate(context.Background(), updatedOrder)
		}
	}()

	return nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orderRepo.GetByIDWithJoins(ctx, id)
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID, status *string, page, limit int) ([]domain.Order, int, error) {
	return s.orderRepo.GetByUserID(ctx, userID, status, page, limit)
}

func (s *OrderService) List(ctx context.Context, f repository.OrderFilter) ([]domain.Order, int, error) {
	return s.orderRepo.List(ctx, f)
}

func (s *OrderService) GetHistory(ctx context.Context, orderID uuid.UUID) ([]domain.OrderHistory, error) {
	return s.orderRepo.GetHistory(ctx, orderID)
}

func (s *OrderService) DecryptCredentials(ctx context.Context, orderID, adminID uuid.UUID) (map[string]string, error) {
	account, err := s.accountRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	login, err := s.encSvc.Decrypt(account.Login)
	if err != nil {
		return nil, errors.New("failed to decrypt credentials")
	}

	password, err := s.encSvc.Decrypt(account.Password)
	if err != nil {
		return nil, errors.New("failed to decrypt credentials")
	}

	s.auditSvc.Log(ctx, adminID, "credentials_decrypt", "order", orderID, map[string]interface{}{
		"account_id": account.ID.String(),
	})

	return map[string]string{
		"login":    string(login),
		"password": string(password),
	}, nil
}

func (s *OrderService) UploadReceipt(ctx context.Context, orderID uuid.UUID, paymentMethod string, receiptURL string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	if order.Status != domain.OrderStatusWaitingPayment {
		return domain.ErrOrderStatusInvalid
	}

	order.PaymentMethod = &paymentMethod
	order.PaymentReceiptURL = &receiptURL
	s.orderRepo.Update(ctx, order)

	return s.changeStatus(ctx, order, domain.OrderStatusPaymentVerification, nil, domain.ChangedByUser, "Receipt uploaded")
}

// Expire2FACodes - timeout check for 2FA codes
func (s *OrderService) Expire2FACodes(ctx context.Context) {
	orders, err := s.orderRepo.GetExpired2FA(ctx, 10*time.Minute)
	if err != nil {
		return
	}
	for _, order := range orders {
		order.AssignedAdminID = nil
		s.orderRepo.Update(ctx, &order)
		s.changeStatus(ctx, &order, domain.OrderStatusCredentialsReceived, nil, domain.ChangedBySystem, "2FA code expired")
	}
}

// ExpireUnpaidOrders - auto-cancel unpaid orders after 24h
func (s *OrderService) ExpireUnpaidOrders(ctx context.Context) {
	orders, err := s.orderRepo.GetExpiredWaitingPayment(ctx, 24*time.Hour)
	if err != nil {
		return
	}
	for _, order := range orders {
		reason := "Auto-cancelled: payment timeout (24h)"
		order.CancelledReason = &reason
		s.orderRepo.Update(ctx, &order)
		s.changeStatus(ctx, &order, domain.OrderStatusCancelled, nil, domain.ChangedBySystem, reason)
	}
}

func (s *OrderService) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	return s.orderRepo.GetDashboardStats(ctx)
}
