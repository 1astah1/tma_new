package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type OrderStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetByIDWithJoins(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	List(ctx context.Context, f repository.OrderFilter) ([]domain.Order, int, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, statuses []string, page, limit int) ([]domain.Order, int, error)
	Create(ctx context.Context, o *domain.Order) error
	CreateInTx(tx *sqlx.Tx, o *domain.Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	Update(ctx context.Context, o *domain.Order) error
	AddHistory(ctx context.Context, h *domain.OrderHistory) error
	AddHistoryInTx(tx *sqlx.Tx, h *domain.OrderHistory) error
	GetHistory(ctx context.Context, orderID uuid.UUID) ([]domain.OrderHistory, error)
	GetExpired2FA(ctx context.Context, timeout time.Duration) ([]domain.Order, error)
	GetExpiredWaitingPayment(ctx context.Context, timeout time.Duration) ([]domain.Order, error)
	GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
}

type ProductStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	List(ctx context.Context, f repository.ProductFilter) ([]domain.Product, int, error)
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type KeyStore interface {
	CountAvailable(ctx context.Context, productID uuid.UUID) (int, error)
	AssignAvailableKey(ctx context.Context, productID, orderID uuid.UUID) (*domain.ProductKey, error)
	ReleaseKey(ctx context.Context, keyID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductKey, error)
}

type AccountStore interface {
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.UserAccount, error)
	Create(ctx context.Context, a *domain.UserAccount) error
	Update(ctx context.Context, a *domain.UserAccount) error
	Update2FACode(ctx context.Context, id uuid.UUID, code *string) error
}

type UserStore interface {
	Upsert(ctx context.Context, telegramID int64, username, firstName *string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	List(ctx context.Context, search string, page, limit int) ([]domain.User, int, error)
	UpdateBan(ctx context.Context, id uuid.UUID, isBanned bool) error
	UpdateAdminNotes(ctx context.Context, id uuid.UUID, notes string) error
}

type ChatStore interface {
	Create(ctx context.Context, msg *domain.ChatMessage) error
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.ChatMessage, error)
}

type Notifier interface {
	SendOrderStatusUpdate(ctx context.Context, order *domain.Order)
}

type Auditor interface {
	Log(ctx context.Context, adminID uuid.UUID, action, targetType string, targetID uuid.UUID, details interface{})
}

type Encryptor interface {
	Encrypt(data []byte) (string, error)
	Decrypt(encoded string) ([]byte, error)
}

type AdminStore interface {
	GetByTelegramID(ctx context.Context, tgID int64) (*domain.Admin, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
	List(ctx context.Context) ([]domain.Admin, error)
	Create(ctx context.Context, a *domain.Admin) error
	Update(ctx context.Context, a *domain.Admin) error
	AddLog(ctx context.Context, log *domain.AdminActionLog) error
	GetLogs(ctx context.Context, f repository.AuditFilter) ([]domain.AdminActionLog, int, error)
}
