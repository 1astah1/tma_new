package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return err
	}
	if log.Details == nil {
		detailsJSON = []byte("{}")
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO admin_actions_log (admin_id, action_type, target_type, target_id, details, ip_address)
		 VALUES ($1, $2, $3, $4, $5::jsonb, $6)`,
		log.AdminID, log.ActionType, log.TargetType, log.TargetID, detailsJSON, log.IPAddress,
	)
	return err
}

type adminActionLogRow struct {
	domain.AdminActionLog
	AdminUsername *string `db:"admin_username"`
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

func (r *AdminRepo) buildAuditLogQuery(f AuditFilter, countOnly bool) (string, []interface{}) {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if f.AdminID != nil {
		where = append(where, fmt.Sprintf("l.admin_id = $%d", argIdx))
		args = append(args, *f.AdminID)
		argIdx++
	}
	if f.ActionType != nil && strings.TrimSpace(*f.ActionType) != "" {
		where = append(where, fmt.Sprintf("l.action_type = $%d", argIdx))
		args = append(args, strings.TrimSpace(*f.ActionType))
		argIdx++
	}
	if f.TargetType != nil && strings.TrimSpace(*f.TargetType) != "" {
		where = append(where, fmt.Sprintf("l.target_type = $%d", argIdx))
		args = append(args, strings.TrimSpace(*f.TargetType))
		argIdx++
	}
	if f.DateFrom != nil && strings.TrimSpace(*f.DateFrom) != "" {
		where = append(where, fmt.Sprintf("l.created_at >= $%d::timestamptz", argIdx))
		args = append(args, strings.TrimSpace(*f.DateFrom))
		argIdx++
	}
	if f.DateTo != nil && strings.TrimSpace(*f.DateTo) != "" {
		where = append(where, fmt.Sprintf("l.created_at <= $%d::timestamptz", argIdx))
		args = append(args, strings.TrimSpace(*f.DateTo))
		argIdx++
	}

	whereSQL := strings.Join(where, " AND ")
	if countOnly {
		return fmt.Sprintf("SELECT COUNT(*) FROM admin_actions_log l WHERE %s", whereSQL), args
	}

	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 50
	}
	offset := (f.Page - 1) * f.Limit

	query := fmt.Sprintf(`
SELECT l.id, l.admin_id, l.action_type, l.target_type, l.target_id, l.details, l.ip_address, l.created_at,
       a.username AS admin_username
FROM admin_actions_log l
LEFT JOIN admins a ON a.id = l.admin_id
WHERE %s
ORDER BY l.created_at DESC
LIMIT $%d OFFSET $%d`, whereSQL, argIdx, argIdx+1)
	args = append(args, f.Limit, offset)
	return query, args
}

func (r *AdminRepo) GetLogs(ctx context.Context, f AuditFilter) ([]domain.AdminActionLog, int, error) {
	countQuery, countArgs := r.buildAuditLogQuery(f, true)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	listQuery, listArgs := r.buildAuditLogQuery(f, false)
	var rows []adminActionLogRow
	if err := r.db.SelectContext(ctx, &rows, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	logs := make([]domain.AdminActionLog, 0, len(rows))
	for _, row := range rows {
		item := row.AdminActionLog
		if row.AdminUsername != nil && strings.TrimSpace(*row.AdminUsername) != "" {
			item.Admin = &domain.Admin{Username: strings.TrimSpace(*row.AdminUsername)}
		}
		logs = append(logs, item)
	}
	return logs, total, nil
}
