package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type AccountRepo struct {
	db *sqlx.DB
}

func NewAccountRepo(db *sqlx.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) Create(ctx context.Context, a *domain.UserAccount) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO user_accounts (user_id, order_id, platform, login, password)
		 VALUES (:user_id, :order_id, :platform, :login, :password)`, a)
	return err
}

func (r *AccountRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.UserAccount, error) {
	var a domain.UserAccount
	err := r.db.GetContext(ctx, &a,
		"SELECT * FROM user_accounts WHERE order_id=$1", orderID)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserAccount, error) {
	var accounts []domain.UserAccount
	err := r.db.SelectContext(ctx, &accounts,
		"SELECT * FROM user_accounts WHERE user_id=$1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	if accounts == nil {
		accounts = []domain.UserAccount{}
	}
	return accounts, nil
}

func (r *AccountRepo) Update2FACode(ctx context.Context, id uuid.UUID, code *string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE user_accounts SET two_factor_code=$1 WHERE id=$2", code, id)
	return err
}

func (r *AccountRepo) Update(ctx context.Context, a *domain.UserAccount) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE user_accounts SET login=:login, password=:password, notes=:notes
		 WHERE id=:id`, a)
	return err
}

type SettingsRepo struct {
	db *sqlx.DB
}

func NewSettingsRepo(db *sqlx.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) Get(ctx context.Context, key string) (map[string]interface{}, error) {
	var value string
	err := r.db.GetContext(ctx, &value,
		"SELECT value::text FROM settings WHERE key=$1", key)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"value": value}, nil
}

func (r *SettingsRepo) GetAll(ctx context.Context) (map[string]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]interface{})
	for rows.Next() {
		var key string
		var value interface{}
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		result[key] = map[string]interface{}{"value": value}
	}
	return result, nil
}

func (r *SettingsRepo) Upsert(ctx context.Context, key string, value interface{}) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO settings (key, value, updated_at) 
		 VALUES ($1, $2::jsonb, NOW())
		 ON CONFLICT (key) DO UPDATE SET value=$2::jsonb, updated_at=NOW()`, key, value)
	return err
}
