package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"tma-backend/internal/domain"
	"tma-backend/internal/service/mocks"
)

func newTestOrderService() (*OrderService, *mocks.MockOrderStore, *mocks.MockProductStore, *mocks.MockKeyStore, *mocks.MockAccountStore, *mocks.MockUserStore, *mocks.MockChatStore, *mocks.MockEncryptor, *mocks.MockNotifier, *mocks.MockAuditor) {
	orderRepo := &mocks.MockOrderStore{}
	productRepo := &mocks.MockProductStore{}
	keyRepo := &mocks.MockKeyStore{}
	accountRepo := &mocks.MockAccountStore{}
	userRepo := &mocks.MockUserStore{}
	chatRepo := &mocks.MockChatStore{}
	encSvc := &mocks.MockEncryptor{}
	notifSvc := &mocks.MockNotifier{}
	auditSvc := &mocks.MockAuditor{}

	svc := &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		keyRepo:     keyRepo,
		accountRepo: accountRepo,
		userRepo:    userRepo,
		chatRepo:    chatRepo,
		encSvc:      encSvc,
		notifSvc:    notifSvc,
		auditSvc:    auditSvc,
	}

	return svc, orderRepo, productRepo, keyRepo, accountRepo, userRepo, chatRepo, encSvc, notifSvc, auditSvc
}

func TestConfirmPayment_Success(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, notifSvc, auditSvc := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()
	paymentMethod := "sbp"

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPaymentVerification,
		UserID: uuid.New(),
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	orderRepo.On("GetByIDWithJoins", ctx, orderID).Return(order, nil)
	orderRepo.On("Update", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusPaid).Return(nil)
	orderRepo.On("AddHistory", ctx, mock.AnythingOfType("*domain.OrderHistory")).Return(nil)
	auditSvc.On("Log", ctx, adminID, "order_status_change", "order", orderID, mock.AnythingOfType("map[string]interface {}")).Return()
	notifSvc.On("SendOrderStatusUpdate", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	err := svc.ConfirmPayment(ctx, orderID, adminID, paymentMethod)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	orderRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestConfirmPayment_OrderNotFound(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return((*domain.Order)(nil), domain.ErrNotFound)

	err := svc.ConfirmPayment(ctx, orderID, adminID, "sbp")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotFound, err)

	orderRepo.AssertExpectations(t)
}

func TestConfirmPayment_WrongStatus(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusNew,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	err := svc.ConfirmPayment(ctx, orderID, adminID, "sbp")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderStatusInvalid, err)

	orderRepo.AssertExpectations(t)
}

func TestIssueKey_Success(t *testing.T) {
	svc, orderRepo, _, keyRepo, _, _, chatRepo, _, notifSvc, auditSvc := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()
	keyID := uuid.New()

	order := &domain.Order{
		ID:             orderID,
		Status:         domain.OrderStatusPaid,
		DeliveryMethod: domain.DeliveryMethodKey,
		ProductID:      uuid.New(),
		UserID:         uuid.New(),
	}

	key := &domain.ProductKey{
		ID:  keyID,
		Key: "TEST-KEY-12345",
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	orderRepo.On("GetByIDWithJoins", ctx, orderID).Return(order, nil).Times(2)
	keyRepo.On("AssignAvailableKey", ctx, order.ProductID, orderID).Return(key, nil)
	orderRepo.On("Update", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	chatRepo.On("Create", ctx, mock.AnythingOfType("*domain.ChatMessage")).Return(nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusKeyIssued).Return(nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusCompleted).Return(nil)
	orderRepo.On("AddHistory", ctx, mock.AnythingOfType("*domain.OrderHistory")).Return(nil).Times(2)
	auditSvc.On("Log", ctx, adminID, "order_status_change", "order", orderID, mock.AnythingOfType("map[string]interface {}")).Return()
	notifSvc.On("SendOrderStatusUpdate", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil).Times(2)

	err := svc.IssueKey(ctx, orderID, adminID)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	orderRepo.AssertExpectations(t)
	keyRepo.AssertExpectations(t)
	chatRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestIssueKey_OrderNotFound(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return((*domain.Order)(nil), domain.ErrNotFound)

	err := svc.IssueKey(ctx, orderID, adminID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotFound, err)

	orderRepo.AssertExpectations(t)
}

func TestIssueKey_WrongStatus(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	order := &domain.Order{
		ID:             orderID,
		Status:         domain.OrderStatusNew,
		DeliveryMethod: domain.DeliveryMethodKey,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	err := svc.IssueKey(ctx, orderID, adminID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderStatusInvalid, err)

	orderRepo.AssertExpectations(t)
}

func TestIssueKey_WrongDeliveryMethod(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	order := &domain.Order{
		ID:             orderID,
		Status:         domain.OrderStatusPaid,
		DeliveryMethod: domain.DeliveryMethodActivation,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	err := svc.IssueKey(ctx, orderID, adminID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderStatusInvalid, err)

	orderRepo.AssertExpectations(t)
}

func TestCancelOrder_ReleasesKey(t *testing.T) {
	svc, orderRepo, _, keyRepo, _, _, _, _, notifSvc, auditSvc := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()
	keyID := uuid.New()
	reason := "Customer request"

	order := &domain.Order{
		ID:      orderID,
		Status:  domain.OrderStatusPaid,
		KeyID:   &keyID,
		UserID:  uuid.New(),
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	orderRepo.On("GetByIDWithJoins", ctx, orderID).Return(order, nil)
	keyRepo.On("ReleaseKey", ctx, keyID).Return(nil)
	orderRepo.On("Update", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusCancelled).Return(nil)
	orderRepo.On("AddHistory", ctx, mock.AnythingOfType("*domain.OrderHistory")).Return(nil)
	auditSvc.On("Log", ctx, adminID, "order_status_change", "order", orderID, mock.AnythingOfType("map[string]interface {}")).Return()
	notifSvc.On("SendOrderStatusUpdate", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	err := svc.CancelOrder(ctx, orderID, adminID, reason)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	orderRepo.AssertExpectations(t)
	keyRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestCancelOrder_NoKey(t *testing.T) {
	svc, orderRepo, _, keyRepo, _, _, _, _, notifSvc, auditSvc := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()
	reason := "Customer request"

	order := &domain.Order{
		ID:      orderID,
		Status:  domain.OrderStatusWaitingPayment,
		KeyID:   nil,
		UserID:  uuid.New(),
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	orderRepo.On("GetByIDWithJoins", ctx, orderID).Return(order, nil)
	orderRepo.On("Update", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusCancelled).Return(nil)
	orderRepo.On("AddHistory", ctx, mock.AnythingOfType("*domain.OrderHistory")).Return(nil)
	auditSvc.On("Log", ctx, adminID, "order_status_change", "order", orderID, mock.AnythingOfType("map[string]interface {}")).Return()
	notifSvc.On("SendOrderStatusUpdate", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	err := svc.CancelOrder(ctx, orderID, adminID, reason)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	orderRepo.AssertExpectations(t)
	keyRepo.AssertNotCalled(t, "ReleaseKey", mock.Anything, mock.Anything)
	auditSvc.AssertExpectations(t)
}

func TestSendChatMessage_Success(t *testing.T) {
	svc, orderRepo, _, _, _, _, chatRepo, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	senderID := uuid.New()
	message := "Hello, how can I help?"

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPaid,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	chatRepo.On("Create", ctx, mock.AnythingOfType("*domain.ChatMessage")).Return(nil)

	err := svc.SendChatMessage(ctx, orderID, senderID, "admin", message)
	require.NoError(t, err)

	orderRepo.AssertExpectations(t)
	chatRepo.AssertExpectations(t)
}

func TestSendChatMessage_OrderNotFound(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	senderID := uuid.New()

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return((*domain.Order)(nil), domain.ErrNotFound)

	err := svc.SendChatMessage(ctx, orderID, senderID, "admin", "test message")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotFound, err)

	orderRepo.AssertExpectations(t)
}

func TestSendChatMessage_WrongStatus(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	senderID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusNew,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	err := svc.SendChatMessage(ctx, orderID, senderID, "admin", "test message")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderStatusInvalid, err)

	orderRepo.AssertExpectations(t)
}

func TestGetChatMessages(t *testing.T) {
	svc, _, _, _, _, _, chatRepo, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	messages := []domain.ChatMessage{
		{ID: uuid.New(), OrderID: orderID, SenderType: "admin", Message: "Hello"},
		{ID: uuid.New(), OrderID: orderID, SenderType: "user", Message: "Hi"},
	}

	ctx := context.Background()
	chatRepo.On("GetByOrderID", ctx, orderID).Return(messages, nil)

	result, err := svc.GetChatMessages(ctx, orderID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Hello", result[0].Message)
	assert.Equal(t, "Hi", result[1].Message)

	chatRepo.AssertExpectations(t)
}

func TestRequestRefund_Success(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, notifSvc, auditSvc := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPaid,
		UserID: uuid.New(),
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	orderRepo.On("GetByIDWithJoins", ctx, orderID).Return(order, nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusRefundRequested).Return(nil)
	orderRepo.On("AddHistory", ctx, mock.AnythingOfType("*domain.OrderHistory")).Return(nil)
	auditSvc.On("Log", ctx, adminID, "order_status_change", "order", orderID, mock.AnythingOfType("map[string]interface {}")).Return()
	notifSvc.On("SendOrderStatusUpdate", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	err := svc.RequestRefund(ctx, orderID, adminID)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	orderRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestRequestRefund_NotRefundable(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusNew,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	err := svc.RequestRefund(ctx, orderID, adminID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderStatusInvalid, err)

	orderRepo.AssertExpectations(t)
}

func TestProcessRefund_ReleasesKey(t *testing.T) {
	svc, orderRepo, _, keyRepo, _, _, _, _, notifSvc, auditSvc := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()
	keyID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusRefundRequested,
		KeyID:  &keyID,
		UserID: uuid.New(),
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	orderRepo.On("GetByIDWithJoins", ctx, orderID).Return(order, nil)
	keyRepo.On("ReleaseKey", ctx, keyID).Return(nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusRefunded).Return(nil)
	orderRepo.On("AddHistory", ctx, mock.AnythingOfType("*domain.OrderHistory")).Return(nil)
	auditSvc.On("Log", ctx, adminID, "order_status_change", "order", orderID, mock.AnythingOfType("map[string]interface {}")).Return()
	notifSvc.On("SendOrderStatusUpdate", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	err := svc.ProcessRefund(ctx, orderID, adminID)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	orderRepo.AssertExpectations(t)
	keyRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestProcessRefund_WrongStatus(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	adminID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPaid,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	err := svc.ProcessRefund(ctx, orderID, adminID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderStatusInvalid, err)

	orderRepo.AssertExpectations(t)
}

func TestCancelOrderByUser_Success(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, notifSvc, _ := newTestOrderService()

	orderID := uuid.New()
	reason := "Changed my mind"

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusNew,
		UserID: uuid.New(),
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	orderRepo.On("GetByIDWithJoins", ctx, orderID).Return(order, nil)
	orderRepo.On("Update", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	orderRepo.On("UpdateStatus", ctx, orderID, domain.OrderStatusCancelled).Return(nil)
	orderRepo.On("AddHistory", ctx, mock.AnythingOfType("*domain.OrderHistory")).Return(nil)
	notifSvc.On("SendOrderStatusUpdate", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	err := svc.CancelOrderByUser(ctx, orderID, reason)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	orderRepo.AssertExpectations(t)
	notifSvc.AssertExpectations(t)
}

func TestCancelOrderByUser_TooLate(t *testing.T) {
	svc, orderRepo, _, _, _, _, _, _, _, _ := newTestOrderService()

	orderID := uuid.New()
	reason := "Changed my mind"

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPaid,
	}

	ctx := context.Background()
	orderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	err := svc.CancelOrderByUser(ctx, orderID, reason)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderStatusInvalid, err)

	orderRepo.AssertExpectations(t)
}
