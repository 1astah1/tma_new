package repository

import (
	"context"

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
	query += " LIMIT $" + string(rune('0'+argIdx)) + " OFFSET $" + string(rune('0'+argIdx+1))
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
