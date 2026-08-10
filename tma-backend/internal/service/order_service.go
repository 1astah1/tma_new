package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type OrderService struct {
	db          *sqlx.DB
	orderRepo   OrderStore
	productRepo ProductStore
	keyRepo     KeyStore
	accountRepo AccountStore
	userRepo    UserStore
	chatRepo    ChatStore
	promoSvc    *PromoService
	encSvc      Encryptor
	notifSvc    Notifier
	auditSvc    Auditor
}

func NewOrderService(
	db *sqlx.DB,
	orderRepo OrderStore,
	productRepo ProductStore,
	keyRepo KeyStore,
	accountRepo AccountStore,
	userRepo UserStore,
	chatRepo ChatStore,
	promoSvc *PromoService,
	encSvc Encryptor,
	notifSvc Notifier,
	auditSvc Auditor,
) *OrderService {
	return &OrderService{
		db: db, orderRepo: orderRepo, productRepo: productRepo,
		keyRepo: keyRepo, accountRepo: accountRepo, userRepo: userRepo, chatRepo: chatRepo,
		promoSvc: promoSvc, encSvc: encSvc, notifSvc: notifSvc, auditSvc: auditSvc,
	}
}

func (s *OrderService) assertActiveUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	if user.IsBanned {
		return domain.ErrForbidden
	}
	return nil
}

func (s *OrderService) assertOrderOwner(ctx context.Context, orderID, userID uuid.UUID) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return order, nil
}

func (s *OrderService) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*domain.Order, error) {
	if err := s.assertActiveUser(ctx, userID); err != nil {
		return nil, err
	}
	order, err := s.orderRepo.GetByIDWithJoins(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return order, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, productID uuid.UUID, deliveryMethod domain.DeliveryMethod, variantID *string, quantity int) (*domain.Order, error) {
	if err := s.assertActiveUser(ctx, userID); err != nil {
		return nil, err
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if product.Status != domain.ProductStatusActive {
		return nil, domain.ErrInvalidInput
	}

	// Все заказы обрабатываются через чат с менеджером.
	deliveryMethod = domain.DeliveryMethodActivation

	if quantity < 1 {
		quantity = 1
	}

	var unitPrice float64 = product.Price

	// Apply variant price if selected
	if variantID != nil && len(product.Variants) > 0 {
		type Variant struct {
			ID    string  `json:"id"`
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}
		var variants []Variant
		if err := json.Unmarshal(product.Variants, &variants); err == nil {
			for _, v := range variants {
				if v.ID == *variantID {
					unitPrice = v.Price
					break
				}
			}
		}
	}

	// Apply discount
	if product.DiscountPercent > 0 {
		unitPrice = unitPrice * (1 - float64(product.DiscountPercent)/100)
	}

	totalAmount := unitPrice * float64(quantity)

	order := &domain.Order{
		UserID:         userID,
		ProductID:      productID,
		VariantID:      variantID,
		Quantity:       quantity,
		DeliveryMethod: deliveryMethod,
		Status:         domain.OrderStatusNew,
		PaymentAmount:  &totalAmount,
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.orderRepo.CreateInTx(tx, order); err != nil {
		return nil, err
	}

	// Transition to WAITING_PAYMENT within transaction
	oldStatus := order.Status
	order.Status = domain.OrderStatusWaitingPayment
	if _, err := tx.ExecContext(ctx,
		"UPDATE orders SET status=$1, updated_at=NOW() WHERE id=$2", order.Status, order.ID); err != nil {
		return nil, err
	}

	history := &domain.OrderHistory{
		OrderID:       order.ID,
		OldStatus:     &oldStatus,
		NewStatus:     order.Status,
		ChangedByID:   nil,
		ChangedByType: domain.ChangedBySystem,
		Comment:       strPtr("Order created"),
	}
	if err := s.orderRepo.AddHistoryInTx(tx, history); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

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
	order, err := s.assertOrderOwner(ctx, orderID, userID)
	if err != nil {
		return err
	}

	if order.Status != domain.OrderStatusAwaitingCredentials && order.Status != domain.OrderStatusCredentialsInvalid {
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

	existing, err := s.accountRepo.GetByOrderID(ctx, orderID)
	if err == nil && existing != nil {
		existing.Login = encLogin
		existing.Password = encPassword
		existing.Platform = platform
		if err := s.accountRepo.Update(ctx, existing); err != nil {
			return err
		}
	} else {
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

func (s *OrderService) Receive2FA(ctx context.Context, orderID, userID uuid.UUID, code string) error {
	order, err := s.assertOrderOwner(ctx, orderID, userID)
	if err != nil {
		return err
	}

	if order.Status != domain.OrderStatusAwaiting2FA && order.Status != domain.OrderStatusInvalid2FA {
		return domain.ErrOrderStatusInvalid
	}

	account, err := s.accountRepo.GetByOrderID(ctx, orderID)
	if err == nil {
		encCode, err := s.encSvc.Encrypt([]byte(code))
		if err != nil {
			return err
		}
		codeStr := string(encCode)
		s.accountRepo.Update2FACode(ctx, account.ID, &codeStr)
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

func (s *OrderService) SimpleStatusChange(ctx context.Context, orderID, adminID uuid.UUID, newStatus domain.OrderStatus, comment string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}

	return s.changeStatus(ctx, order, newStatus, &adminID, domain.ChangedByAdmin, comment)
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

func (s *OrderService) CancelOrderByUser(ctx context.Context, orderID, userID uuid.UUID, reason string) error {
	order, err := s.assertOrderOwner(ctx, orderID, userID)
	if err != nil {
		return err
	}

	allowable := map[domain.OrderStatus]bool{
		domain.OrderStatusNew:              true,
		domain.OrderStatusWaitingPayment:   true,
		domain.OrderStatusPaymentVerification: true,
	}
	if !allowable[order.Status] {
		return domain.ErrOrderStatusInvalid
	}

	order.CancelledReason = &reason
	s.orderRepo.Update(ctx, order)

	return s.changeStatus(ctx, order, domain.OrderStatusCancelled, nil, domain.ChangedByUser, reason)
}

func (s *OrderService) RequestRefundByUser(ctx context.Context, orderID, userID uuid.UUID, reason string) error {
	order, err := s.assertOrderOwner(ctx, orderID, userID)
	if err != nil {
		return err
	}

	if !canRefund(order.Status) {
		return domain.ErrOrderStatusInvalid
	}

	return s.changeStatus(ctx, order, domain.OrderStatusRefundRequested, nil, domain.ChangedByUser, reason)
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
	case domain.OrderStatusPaymentVerification, domain.OrderStatusPaid, domain.OrderStatusWaitingActivation,
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

func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID, statuses []string, page, limit int) ([]domain.Order, int, error) {
	return s.orderRepo.GetByUserID(ctx, userID, statuses, page, limit)
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

func (s *OrderService) ApplyPromoToOrders(ctx context.Context, orders []*domain.Order, promoCode string) error {
	if promoCode == "" || len(orders) == 0 || s.promoSvc == nil {
		return nil
	}

	var subtotal float64
	for _, order := range orders {
		if order.PaymentAmount != nil {
			subtotal += *order.PaymentAmount
		}
	}
	if subtotal <= 0 {
		return nil
	}

	_, finalTotal, err := s.promoSvc.ValidateAndApply(promoCode, subtotal)
	if err != nil {
		return err
	}

	discount := subtotal - finalTotal
	if discount <= 0 {
		return s.promoSvc.ApplyPromo(promoCode)
	}

	remaining := discount
	for i, order := range orders {
		if order.PaymentAmount == nil {
			continue
		}
		var orderDiscount float64
		if i == len(orders)-1 {
			orderDiscount = remaining
		} else {
			orderDiscount = discount * (*order.PaymentAmount / subtotal)
			remaining -= orderDiscount
		}
		newAmount := *order.PaymentAmount - orderDiscount
		if newAmount < 0 {
			newAmount = 0
		}
		order.PaymentAmount = &newAmount
		if err := s.orderRepo.Update(ctx, order); err != nil {
			return err
		}
	}

	return s.promoSvc.ApplyPromo(promoCode)
}

func (s *OrderService) UploadReceipt(ctx context.Context, orderID, userID uuid.UUID, paymentMethod string, receiptURL string) error {
	order, err := s.assertOrderOwner(ctx, orderID, userID)
	if err != nil {
		return err
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

// ExpireUnpaidOrders - remind at 20h, auto-cancel unpaid orders after 24h
func (s *OrderService) ExpireUnpaidOrders(ctx context.Context) {
	reminders, err := s.orderRepo.GetWaitingPaymentOlderThan(ctx, 20*time.Hour, 21*time.Hour)
	if err == nil {
		for _, order := range reminders {
			full, err := s.orderRepo.GetByIDWithJoins(ctx, order.ID)
			if err == nil && full.User != nil {
				text := fmt.Sprintf("⏰ Напоминание: заказ #%s ожидает оплаты. Загрузите чек в приложении или свяжитесь с менеджером.", order.ID.String()[:8])
				s.notifSvc.SendUserMessage(ctx, full.User.TelegramID, text)
			}
		}
	}

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

func (s *OrderService) GetAccount(ctx context.Context, orderID uuid.UUID) (*domain.UserAccount, error) {
	return s.accountRepo.GetByOrderID(ctx, orderID)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	return s.orderRepo.UpdateStatus(ctx, id, status)
}

func (s *OrderService) AddHistory(ctx context.Context, h *domain.OrderHistory) error {
	return s.orderRepo.AddHistory(ctx, h)
}

func (s *OrderService) BulkUpdateStatus(ctx context.Context, orderIDs []uuid.UUID, newStatus domain.OrderStatus, adminID uuid.UUID, comment string) (int, error) {
	success := 0
	for _, id := range orderIDs {
		order, err := s.orderRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		if !domain.IsValidTransition(order.Status, newStatus) {
			continue
		}
		if err := s.SimpleStatusChange(ctx, id, adminID, newStatus, comment); err != nil {
			continue
		}
		success++
	}
	return success, nil
}

func (s *OrderService) SendChatMessage(ctx context.Context, orderID, senderID uuid.UUID, senderType string, message string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return domain.ErrNotFound
	}
	if senderType == "user" && order.UserID != senderID {
		return domain.ErrForbidden
	}

	allowable := map[domain.OrderStatus]bool{
		domain.OrderStatusPaid:              true,
		domain.OrderStatusWaitingActivation: true,
		domain.OrderStatusAwaitingCredentials: true,
		domain.OrderStatusCredentialsReceived: true,
		domain.OrderStatusCredentialsInvalid:  true,
		domain.OrderStatusAwaiting2FA:         true,
		domain.OrderStatusInvalid2FA:          true,
		domain.OrderStatusActivating:          true,
		domain.OrderStatusActivated:           true,
		domain.OrderStatusKeyIssued:           true,
		domain.OrderStatusCompleted:           true,
		domain.OrderStatusCancelled:           true,
		domain.OrderStatusRefundRequested:     true,
		domain.OrderStatusRefunded:            true,
	}
	if !allowable[order.Status] {
		return domain.ErrOrderStatusInvalid
	}

	msg := &domain.ChatMessage{
		OrderID:    orderID,
		SenderType: senderType,
		SenderID:   senderID,
		Message:    message,
	}
	return s.chatRepo.Create(ctx, msg)
}

func (s *OrderService) GetChatMessages(ctx context.Context, orderID uuid.UUID) ([]domain.ChatMessage, error) {
	return s.chatRepo.GetByOrderID(ctx, orderID)
}

func (s *OrderService) GetChatMessagesForUser(ctx context.Context, orderID, userID uuid.UUID) ([]domain.ChatMessage, error) {
	if _, err := s.assertOrderOwner(ctx, orderID, userID); err != nil {
		return nil, err
	}
	return s.chatRepo.GetByOrderID(ctx, orderID)
}

func strPtr(s string) *string {
	return &s
}
