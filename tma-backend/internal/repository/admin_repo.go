package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type AdminRepo struct {
	db *sqlx.DB
}

func NewAdminRepo(db *sqlx.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) GetByTelegramID(ctx context.Context, tgID int64) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.GetContext(ctx, &a, "SELECT * FROM admins WHERE telegram_id = $1", tgID)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.GetContext(ctx, &a, "SELECT * FROM admins WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) List(ctx context.Context) ([]domain.Admin, error) {
	var admins []domain.Admin
	err := r.db.SelectContext(ctx, &admins, "SELECT * FROM admins ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	if admins == nil {
		admins = []domain.Admin{}
	}
	return admins, nil
}

func (r *AdminRepo) Create(ctx context.Context, a *domain.Admin) error {
	err := r.db.GetContext(ctx, a,
		`INSERT INTO admins (telegram_id, username, password_hash, roles, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING *`, a.TelegramID, a.Username, a.PasswordHash, a.Roles, a.IsActive)
	return err
}

func (r *AdminRepo) Update(ctx context.Context, a *domain.Admin) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE admins SET username=:username, roles=:roles, is_active=:is_active
		 WHERE id=:id`, a)
	return err
}

func (r *AdminRepo) AddLog(ctx context.Context, log *domain.AdminActionLog) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO admin_actions_log (admin_id, action_type, target_type, target_id, details, ip_address)
		 VALUES (:admin_id, :action_type, :target_type, :target_id, :details, :ip_address)`, log)
	return err
}

type AuditFilter struct {
	AdminID    *uuid.UUID `json:"admin_id,omitempty"`
	ActionType *string    `json:"action_type,omitempty"`
	TargetType *string    `json:"target_type,omitempty"`
	DateFrom   *string    `json:"date_from,omitempty"`
	DateTo     *string    `json:"date_to,omitempty"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

func (r *AdminRepo) GetLogs(ctx context.Context, f AuditFilter) ([]domain.AdminActionLog, int, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM admin_actions_log"
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 50
	}
	offset := (f.Page - 1) * f.Limit

	var logs []domain.AdminActionLog
	err := r.db.SelectContext(ctx, &logs,
		`SELECT l.*, row_to_json(a.*) as admin
		 FROM admin_actions_log l
		 LEFT JOIN admins a ON a.id = l.admin_id
		 ORDER BY l.created_at DESC
		 LIMIT $1 OFFSET $2`, f.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if logs == nil {
		logs = []domain.AdminActionLog{}
	}
	return logs, total, nil
}
