package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type MockOrderStore struct {
	mock.Mock
}

func (m *MockOrderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderStore) GetByIDWithJoins(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderStore) List(ctx context.Context, f repository.OrderFilter) ([]domain.Order, int, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderStore) GetByUserID(ctx context.Context, userID uuid.UUID, statuses []string, page, limit int) ([]domain.Order, int, error) {
	args := m.Called(ctx, userID, statuses, page, limit)
	return args.Get(0).([]domain.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderStore) Create(ctx context.Context, o *domain.Order) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}

func (m *MockOrderStore) CreateInTx(tx *sqlx.Tx, o *domain.Order) error {
	args := m.Called(tx, o)
	return args.Error(0)
}

func (m *MockOrderStore) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderStore) Update(ctx context.Context, o *domain.Order) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}

func (m *MockOrderStore) AddHistory(ctx context.Context, h *domain.OrderHistory) error {
	args := m.Called(ctx, h)
	return args.Error(0)
}

func (m *MockOrderStore) AddHistoryInTx(tx *sqlx.Tx, h *domain.OrderHistory) error {
	args := m.Called(tx, h)
	return args.Error(0)
}

func (m *MockOrderStore) GetHistory(ctx context.Context, orderID uuid.UUID) ([]domain.OrderHistory, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]domain.OrderHistory), args.Error(1)
}

func (m *MockOrderStore) GetExpired2FA(ctx context.Context, timeout time.Duration) ([]domain.Order, error) {
	args := m.Called(ctx, timeout)
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderStore) GetExpiredWaitingPayment(ctx context.Context, timeout time.Duration) ([]domain.Order, error) {
	args := m.Called(ctx, timeout)
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderStore) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

type MockProductStore struct {
	mock.Mock
}

func (m *MockProductStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductStore) List(ctx context.Context, f repository.ProductFilter) ([]domain.Product, int, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.Product), args.Int(1), args.Error(2)
}

func (m *MockProductStore) Create(ctx context.Context, p *domain.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProductStore) Update(ctx context.Context, p *domain.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProductStore) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockKeyStore struct {
	mock.Mock
}

func (m *MockKeyStore) CountAvailable(ctx context.Context, productID uuid.UUID) (int, error) {
	args := m.Called(ctx, productID)
	return args.Int(0), args.Error(1)
}

func (m *MockKeyStore) AssignAvailableKey(ctx context.Context, productID, orderID uuid.UUID) (*domain.ProductKey, error) {
	args := m.Called(ctx, productID, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductKey), args.Error(1)
}

func (m *MockKeyStore) ReleaseKey(ctx context.Context, keyID uuid.UUID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockKeyStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductKey), args.Error(1)
}

type MockAccountStore struct {
	mock.Mock
}

func (m *MockAccountStore) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.UserAccount, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserAccount), args.Error(1)
}

func (m *MockAccountStore) Create(ctx context.Context, a *domain.UserAccount) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockAccountStore) Update(ctx context.Context, a *domain.UserAccount) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockAccountStore) Update2FACode(ctx context.Context, id uuid.UUID, code *string) error {
	args := m.Called(ctx, id, code)
	return args.Error(0)
}

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Upsert(ctx context.Context, telegramID int64, username, firstName *string) (*domain.User, error) {
	args := m.Called(ctx, telegramID, username, firstName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStore) List(ctx context.Context, search string, page, limit int) ([]domain.User, int, error) {
	args := m.Called(ctx, search, page, limit)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *MockUserStore) UpdateBan(ctx context.Context, id uuid.UUID, isBanned bool) error {
	args := m.Called(ctx, id, isBanned)
	return args.Error(0)
}

func (m *MockUserStore) UpdateAdminNotes(ctx context.Context, id uuid.UUID, notes string) error {
	args := m.Called(ctx, id, notes)
	return args.Error(0)
}

type MockChatStore struct {
	mock.Mock
}

func (m *MockChatStore) Create(ctx context.Context, msg *domain.ChatMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockChatStore) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.ChatMessage, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]domain.ChatMessage), args.Error(1)
}

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) {
	m.Called(ctx, order)
}

type MockAuditor struct {
	mock.Mock
}

func (m *MockAuditor) Log(ctx context.Context, adminID uuid.UUID, action, targetType string, targetID uuid.UUID, details interface{}) {
	m.Called(ctx, adminID, action, targetType, targetID, details)
}

type MockEncryptor struct {
	mock.Mock
}

func (m *MockEncryptor) Encrypt(data []byte) (string, error) {
	args := m.Called(data)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockEncryptor) Decrypt(encoded string) ([]byte, error) {
	args := m.Called(encoded)
	return args.Get(0).([]byte), args.Error(1)
}

type MockAdminStore struct {
	mock.Mock
}

func (m *MockAdminStore) GetByTelegramID(ctx context.Context, tgID int64) (*domain.Admin, error) {
	args := m.Called(ctx, tgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Admin), args.Error(1)
}

func (m *MockAdminStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Admin), args.Error(1)
}

func (m *MockAdminStore) List(ctx context.Context) ([]domain.Admin, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Admin), args.Error(1)
}

func (m *MockAdminStore) Create(ctx context.Context, a *domain.Admin) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockAdminStore) Update(ctx context.Context, a *domain.Admin) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockAdminStore) AddLog(ctx context.Context, log *domain.AdminActionLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAdminStore) GetLogs(ctx context.Context, f repository.AuditFilter) ([]domain.AdminActionLog, int, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.AdminActionLog), args.Int(1), args.Error(2)
}
