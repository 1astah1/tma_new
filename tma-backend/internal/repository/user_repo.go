package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByTelegramID(ctx context.Context, tgID int64) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE telegram_id = $1", tgID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, tgID int64, username, firstName *string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u,
		`INSERT INTO users (telegram_id, username, first_name) 
		 VALUES ($1, $2, $3) 
		 RETURNING *`, tgID, username, firstName)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Upsert(ctx context.Context, tgID int64, username, firstName *string) (*domain.User, error) {
	var u domain.User
	now := time.Now()
	err := r.db.GetContext(ctx, &u,
		`INSERT INTO users (telegram_id, username, first_name, last_interaction, updated_at)
		 VALUES ($1, $2, $3, $4, $4)
		 ON CONFLICT (telegram_id) 
		 DO UPDATE SET username = COALESCE($2, users.username), 
		               first_name = COALESCE($3, users.first_name),
		               last_interaction = $4,
		               updated_at = $4
		 RETURNING *`, tgID, username, firstName, now)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) List(ctx context.Context, search, sortField, sortOrder string, page, limit int) ([]domain.User, int, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM users"
	var countArgs []interface{}
	if search != "" {
		countQuery += " WHERE username ILIKE '%' || $1 || '%' OR first_name ILIKE '%' || $1 || '%'"
		countArgs = append(countArgs, search)
	}
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	orderBy := "created_at DESC"
	switch sortField {
	case "last_interaction":
		if strings.EqualFold(sortOrder, "ASC") {
			orderBy = "last_interaction ASC NULLS LAST"
		} else {
			orderBy = "last_interaction DESC NULLS LAST"
		}
	case "created_at":
		if strings.EqualFold(sortOrder, "ASC") {
			orderBy = "created_at ASC"
		} else {
			orderBy = "created_at DESC"
		}
	}

	offset := (page - 1) * limit
	query := "SELECT * FROM users"
	var dataArgs []interface{}
	if search != "" {
		query += " WHERE username ILIKE '%' || $1 || '%' OR first_name ILIKE '%' || $1 || '%'"
		dataArgs = append(dataArgs, search)
	}
	query += fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", orderBy, len(dataArgs)+1, len(dataArgs)+2)
	dataArgs = append(dataArgs, limit, offset)

	var users []domain.User
	err := r.db.SelectContext(ctx, &users, query, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	if users == nil {
		users = []domain.User{}
	}
	return users, total, nil
}

func (r *UserRepo) UpdateBan(ctx context.Context, id uuid.UUID, isBanned bool) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET is_banned = $1 WHERE id = $2", isBanned, id)
	return err
}

func (r *UserRepo) UpdateAdminNotes(ctx context.Context, id uuid.UUID, notes string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET admin_notes = $1 WHERE id = $2", notes, id)
	return err
}

func (r *UserRepo) CountOrders(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID)
	return count, err
}

func (r *UserRepo) CountOrdersByStatus(ctx context.Context, userID uuid.UUID, status string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status = $2", userID, status)
	return count, err
}

func (r *UserRepo) ListTelegramIDsForBroadcast(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := r.db.SelectContext(ctx, &ids,
		`SELECT telegram_id FROM users WHERE is_banned = false AND telegram_id > 0 ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []int64{}
	}
	return ids, nil
}
