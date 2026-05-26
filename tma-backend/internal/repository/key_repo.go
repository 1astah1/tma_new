package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type KeyRepo struct {
	db *sqlx.DB
}

func NewKeyRepo(db *sqlx.DB) *KeyRepo {
	return &KeyRepo{db: db}
}

func (r *KeyRepo) AssignAvailableKey(ctx context.Context, productID, orderID uuid.UUID) (*domain.ProductKey, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var key domain.ProductKey
	err = tx.GetContext(ctx, &key,
		`SELECT * FROM product_keys 
		 WHERE product_id = $1 AND status = 'available' 
		 ORDER BY created_at ASC LIMIT 1 
		 FOR UPDATE SKIP LOCKED`, productID)
	if err != nil {
		return nil, domain.ErrKeyNotAvailable
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE product_keys SET status='sold', order_id=$1 WHERE id=$2",
		orderID, key.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	key.Status = domain.KeyStatusSold
	key.OrderID = &orderID
	return &key, nil
}

func (r *KeyRepo) GetByProductID(ctx context.Context, productID uuid.UUID, status *string, page, limit int) ([]domain.ProductKey, int, error) {
	args := []interface{}{productID}
	where := "WHERE product_id = $1"
	argIdx := 2

	if status != nil && *status != "" {
		where += " AND status = $2"
		args = append(args, *status)
		argIdx++
	}

	var total int
	if err := r.db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM product_keys "+where, args[:argIdx-1]...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := "SELECT * FROM product_keys " + where + " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var keys []domain.ProductKey
	if err := r.db.SelectContext(ctx, &keys, query, args...); err != nil {
		return nil, 0, err
	}
	if keys == nil {
		keys = []domain.ProductKey{}
	}
	return keys, total, nil
}

func (r *KeyRepo) BulkImport(ctx context.Context, productID uuid.UUID, keys []string) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	imported := 0
	for _, k := range keys {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO product_keys (product_id, key) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			productID, k)
		if err == nil {
			imported++
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return imported, nil
}

func (r *KeyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductKey, error) {
	var k domain.ProductKey
	err := r.db.GetContext(ctx, &k, "SELECT * FROM product_keys WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *KeyRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.ProductKey, error) {
	var k domain.ProductKey
	err := r.db.GetContext(ctx, &k, "SELECT * FROM product_keys WHERE order_id=$1", orderID)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *KeyRepo) ListAll(ctx context.Context, status *string, page, limit int) ([]domain.ProductKey, int, error) {
	args := []interface{}{}
	where := ""
	argIdx := 1

	if status != nil && *status != "" {
		where = "WHERE status = $1"
		args = append(args, *status)
		argIdx++
	}

	var total int
	query := "SELECT COUNT(*) FROM product_keys " + where
	if err := r.db.GetContext(ctx, &total, query, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query = "SELECT * FROM product_keys " + where + " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var keys []domain.ProductKey
	if err := r.db.SelectContext(ctx, &keys, query, args...); err != nil {
		return nil, 0, err
	}
	if keys == nil {
		keys = []domain.ProductKey{}
	}
	return keys, total, nil
}

func (r *KeyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM product_keys WHERE id=$1", id)
	return err
}

func (r *KeyRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE product_keys SET status=$1 WHERE id=$2", status, id)
	return err
}

func (r *KeyRepo) BulkDelete(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query := "DELETE FROM product_keys WHERE id = ANY($1)"
	res, err := r.db.ExecContext(ctx, query, ids)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *KeyRepo) AssignSpecificKey(ctx context.Context, keyID, orderID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"UPDATE product_keys SET status='sold', order_id=$1 WHERE id=$2 AND status='available'",
		orderID, keyID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *KeyRepo) CreateKey(ctx context.Context, productID uuid.UUID, key string, status domain.KeyStatus) (*domain.ProductKey, error) {
	var k domain.ProductKey
	err := r.db.GetContext(ctx, &k,
		"INSERT INTO product_keys (product_id, key, status) VALUES ($1, $2, $3) RETURNING *",
		productID, key, status)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *KeyRepo) CountAvailable(ctx context.Context, productID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM product_keys WHERE product_id = $1 AND status = 'available'", productID)
	return count, err
}

func (r *KeyRepo) ReleaseKey(ctx context.Context, keyID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE product_keys SET status = 'available', order_id = NULL WHERE id = $1 AND status = 'sold'", keyID)
	return err
}

type KeyStats struct {
	Available int `json:"available"`
	Sold      int `json:"sold"`
	Reserved  int `json:"reserved"`
	Invalid   int `json:"invalid"`
	Total     int `json:"total"`
}

func (r *KeyRepo) GetStatsByProductID(ctx context.Context, productID uuid.UUID) (*KeyStats, error) {
	var stats KeyStats
	err := r.db.GetContext(ctx, &stats, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'available') as available,
			COUNT(*) FILTER (WHERE status = 'sold') as sold,
			COUNT(*) FILTER (WHERE status = 'reserved') as reserved,
			COUNT(*) FILTER (WHERE status = 'invalid') as invalid,
			COUNT(*) as total
		FROM product_keys WHERE product_id = $1`, productID)
	return &stats, err
}

func (r *KeyRepo) UpdateKey(ctx context.Context, id uuid.UUID, key string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE product_keys SET key = $1 WHERE id = $2", key, id)
	return err
}
