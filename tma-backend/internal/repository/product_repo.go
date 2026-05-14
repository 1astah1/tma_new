package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type ProductRepo struct {
	db *sqlx.DB
}

func NewProductRepo(db *sqlx.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

type ProductFilter struct {
	Platform *string  `json:"platform,omitempty"`
	Type     *string  `json:"type,omitempty"`
	Search   *string  `json:"search,omitempty"`
	MinPrice *float64 `json:"min_price,omitempty"`
	MaxPrice *float64 `json:"max_price,omitempty"`
	Status   *string  `json:"status,omitempty"`
	Page     int      `json:"page"`
	Limit    int      `json:"limit"`
}

func (r *ProductRepo) List(ctx context.Context, f ProductFilter) ([]domain.Product, int, error) {
	args := []interface{}{}
	where := []string{}

	argIdx := 1
	if f.Platform != nil && *f.Platform != "" {
		where = append(where, fmt.Sprintf("platform = $%d", argIdx))
		args = append(args, *f.Platform)
		argIdx++
	}
	if f.Type != nil && *f.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *f.Type)
		argIdx++
	}
	if f.Search != nil && *f.Search != "" {
		where = append(where, fmt.Sprintf("title ILIKE '%%' || $%d || '%%'", argIdx))
		args = append(args, *f.Search)
		argIdx++
	}
	if f.MinPrice != nil {
		where = append(where, fmt.Sprintf("price >= $%d", argIdx))
		args = append(args, *f.MinPrice)
		argIdx++
	}
	if f.MaxPrice != nil {
		where = append(where, fmt.Sprintf("price <= $%d", argIdx))
		args = append(args, *f.MaxPrice)
		argIdx++
	}
	if f.Status != nil && *f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM products" + whereClause
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	query := "SELECT * FROM products" + whereClause + " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, f.Limit, offset)

	var products []domain.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, 0, err
	}
	if products == nil {
		products = []domain.Product{}
	}
	return products, total, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var p domain.Product
	err := r.db.GetContext(ctx, &p, "SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	err := r.db.GetContext(ctx, p,
		`INSERT INTO products (title, description, platform, type, price, image_url, delivery_methods, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING *`, p.Title, p.Description, p.Platform, p.Type, p.Price, p.ImageURL, p.DeliveryMethods, p.Status)
	return err
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE products SET title=:title, description=:description, platform=:platform, 
		 type=:type, price=:price, image_url=:image_url, 
		 delivery_methods=:delivery_methods, status=:status, updated_at=NOW()
		 WHERE id=:id`, p)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE products SET status='inactive' WHERE id=$1", id)
	return err
}
