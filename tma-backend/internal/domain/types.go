package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Platform string
type ProductType string
type ProductStatus string
type DeliveryMethod string
type PaymentMethod string
type OrderStatus string
type KeyStatus string
type ChangedByType string
type TransactionStatus string

const (
	PlatformPS4  Platform = "ps4"
	PlatformPS5  Platform = "ps5"
	PlatformXbox Platform = "xbox"

	ProductTypeGame         ProductType = "game"
	ProductTypeCurrency     ProductType = "currency"
	ProductTypeSubscription ProductType = "subscription"

	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"

	DeliveryMethodKey        DeliveryMethod = "key"
	DeliveryMethodActivation DeliveryMethod = "activation"

	PaymentMethodSBP    PaymentMethod = "sbp"
	PaymentMethodCard   PaymentMethod = "card"
	PaymentMethodCrypto PaymentMethod = "crypto"
	PaymentMethodStars  PaymentMethod = "telegram_stars"

	KeyStatusAvailable KeyStatus = "available"
	KeyStatusReserved  KeyStatus = "reserved"
	KeyStatusSold      KeyStatus = "sold"

	ChangedByAdmin ChangedByType = "admin"
	ChangedBySystem ChangedByType = "system"
	ChangedByUser  ChangedByType = "user"

	TransactionPending  TransactionStatus = "pending"
	TransactionVerified TransactionStatus = "verified"
	TransactionFailed   TransactionStatus = "failed"
)

// Order statuses
const (
	OrderStatusNew                OrderStatus = "NEW"
	OrderStatusWaitingPayment     OrderStatus = "WAITING_PAYMENT"
	OrderStatusPaymentVerification OrderStatus = "PAYMENT_VERIFICATION"
	OrderStatusPaid               OrderStatus = "PAID"
	OrderStatusWaitingActivation  OrderStatus = "WAITING_ACTIVATION"
	OrderStatusAwaitingCredentials OrderStatus = "AWAITING_CREDENTIALS"
	OrderStatusCredentialsReceived OrderStatus = "CREDENTIALS_RECEIVED"
	OrderStatusAwaiting2FA        OrderStatus = "AWAITING_2FA"
	OrderStatusActivating         OrderStatus = "ACTIVATING"
	OrderStatusActivated          OrderStatus = "ACTIVATED"
	OrderStatusKeyIssued          OrderStatus = "KEY_ISSUED"
	OrderStatusCompleted          OrderStatus = "COMPLETED"
	OrderStatusCancelled          OrderStatus = "CANCELLED"
	OrderStatusRefundRequested    OrderStatus = "REFUND_REQUESTED"
	OrderStatusRefunded           OrderStatus = "REFUNDED"
)

type User struct {
	ID              uuid.UUID `db:"id" json:"id"`
	TelegramID      int64     `db:"telegram_id" json:"telegram_id"`
	Username        *string   `db:"username" json:"username"`
	FirstName       *string   `db:"first_name" json:"first_name"`
	LastInteraction *time.Time `db:"last_interaction" json:"last_interaction"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type Product struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	Title           string          `db:"title" json:"title"`
	Description     *string         `db:"description" json:"description"`
	Platform        Platform        `db:"platform" json:"platform"`
	Type            ProductType     `db:"type" json:"type"`
	Price           float64         `db:"price" json:"price"`
	ImageURL        *string         `db:"image_url" json:"image_url"`
	DeliveryMethods pq.StringArray  `db:"delivery_methods" json:"delivery_methods"`
	Status          ProductStatus   `db:"status" json:"status"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

type ProductKey struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ProductID uuid.UUID `db:"product_id" json:"product_id"`
	Key       string    `db:"key" json:"-"`
	Status    KeyStatus `db:"status" json:"status"`
	OrderID   *uuid.UUID `db:"order_id" json:"order_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Order struct {
	ID                uuid.UUID     `db:"id" json:"id"`
	UserID            uuid.UUID     `db:"user_id" json:"user_id"`
	ProductID         uuid.UUID     `db:"product_id" json:"product_id"`
	DeliveryMethod    DeliveryMethod `db:"delivery_method" json:"delivery_method"`
	Status            OrderStatus   `db:"status" json:"status"`
	PaymentMethod     *string       `db:"payment_method" json:"payment_method"`
	PaymentAmount     *float64      `db:"payment_amount" json:"payment_amount"`
	PaymentReceiptURL *string       `db:"payment_receipt_url" json:"payment_receipt_url"`
	PaymentVerifiedBy *uuid.UUID    `db:"payment_verified_by" json:"payment_verified_by"`
	KeyID             *uuid.UUID    `db:"key_id" json:"key_id"`
	AssignedAdminID   *uuid.UUID    `db:"assigned_admin_id" json:"assigned_admin_id"`
	CancelledReason   *string       `db:"cancelled_reason" json:"cancelled_reason"`
	CreatedAt         time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at" json:"updated_at"`
	// Joined fields
	Product    *Product `db:"-" json:"product,omitempty"`
	User       *User    `db:"-" json:"user,omitempty"`
}

type OrderHistory struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	OrderID       uuid.UUID      `db:"order_id" json:"order_id"`
	OldStatus     *OrderStatus   `db:"old_status" json:"old_status"`
	NewStatus     OrderStatus    `db:"new_status" json:"new_status"`
	ChangedByID   *uuid.UUID     `db:"changed_by_id" json:"changed_by_id"`
	ChangedByType ChangedByType  `db:"changed_by_type" json:"changed_by_type"`
	Comment       *string        `db:"comment" json:"comment"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

type UserAccount struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	OrderID       uuid.UUID `db:"order_id" json:"order_id"`
	Platform      Platform  `db:"platform" json:"platform"`
	Login         string    `db:"login" json:"-"`
	Password      string    `db:"password" json:"-"`
	TwoFactorCode *string   `db:"two_factor_code" json:"-"`
	Notes         *string   `db:"notes" json:"notes"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type Admin struct {
	ID           uuid.UUID `db:"id" json:"id"`
	TelegramID   int64     `db:"telegram_id" json:"telegram_id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash *string   `db:"password_hash" json:"-"`
	Roles        pq.StringArray  `db:"roles" json:"roles"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type AdminActionLog struct {
	ID         uuid.UUID              `db:"id" json:"id"`
	AdminID    uuid.UUID              `db:"admin_id" json:"admin_id"`
	ActionType string                 `db:"action_type" json:"action_type"`
	TargetType string                 `db:"target_type" json:"target_type"`
	TargetID   *uuid.UUID             `db:"target_id" json:"target_id"`
	Details    map[string]interface{} `db:"details" json:"details"`
	IPAddress  *string                `db:"ip_address" json:"ip_address"`
	CreatedAt  time.Time              `db:"created_at" json:"created_at"`
	Admin      *Admin                 `db:"-" json:"admin,omitempty"`
}

type PaymentTransaction struct {
	ID         uuid.UUID         `db:"id" json:"id"`
	OrderID    uuid.UUID         `db:"order_id" json:"order_id"`
	Method     string            `db:"method" json:"method"`
	Amount     float64           `db:"amount" json:"amount"`
	Currency   string            `db:"currency" json:"currency"`
	ReceiptURL *string           `db:"receipt_url" json:"receipt_url"`
	Status     TransactionStatus `db:"status" json:"status"`
	VerifiedBy *uuid.UUID        `db:"verified_by" json:"verified_by"`
	CreatedAt  time.Time         `db:"created_at" json:"created_at"`
}
