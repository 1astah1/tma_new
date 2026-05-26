package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PromoCode struct {
	ID              uuid.UUID  `db:"id"`
	Code            string     `db:"code"`
	DiscountPercent float64    `db:"discount_percent"`
	DiscountFixed   float64    `db:"discount_fixed"`
	UsageLimit      *int       `db:"usage_limit"`
	UsedCount       int        `db:"used_count"`
	ValidFrom       time.Time  `db:"valid_from"`
	ValidUntil      *time.Time `db:"valid_until"`
	IsActive        bool       `db:"is_active"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type PromoCodeRepo struct {
	db *sqlx.DB
}

func NewPromoCodeRepo(db *sqlx.DB) *PromoCodeRepo {
	return &PromoCodeRepo{db: db}
}

func (r *PromoCodeRepo) GetByCode(code string) (*PromoCode, error) {
	var promo PromoCode
	err := r.db.Get(&promo, "SELECT * FROM promo_codes WHERE code = $1 AND is_active = TRUE", code)
	if err != nil {
		return nil, err
	}
	return &promo, nil
}

func (r *PromoCodeRepo) IsValid(promo *PromoCode) bool {
	if !promo.IsActive {
		return false
	}
	if promo.ValidUntil != nil && time.Now().After(*promo.ValidUntil) {
		return false
	}
	if time.Now().Before(promo.ValidFrom) {
		return false
	}
	if promo.UsageLimit != nil && promo.UsedCount >= *promo.UsageLimit {
		return false
	}
	return true
}

func (r *PromoCodeRepo) IncrementUsage(id uuid.UUID) error {
	_, err := r.db.Exec("UPDATE promo_codes SET used_count = used_count + 1, updated_at = NOW() WHERE id = $1", id)
	return err
}

func (r *PromoCodeRepo) List() ([]PromoCode, error) {
	var promos []PromoCode
	err := r.db.Select(&promos, "SELECT * FROM promo_codes ORDER BY created_at DESC")
	return promos, err
}

func (r *PromoCodeRepo) Create(promo *PromoCode) error {
	return r.db.Get(promo, `
		INSERT INTO promo_codes (code, discount_percent, discount_fixed, usage_limit, valid_from, valid_until, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`, promo.Code, promo.DiscountPercent, promo.DiscountFixed, promo.UsageLimit, promo.ValidFrom, promo.ValidUntil, promo.IsActive)
}

func (r *PromoCodeRepo) Update(id uuid.UUID, updates map[string]interface{}) error {
	setClauses := []string{}
	args := []interface{}{}
	i := 1
	for k, v := range updates {
		setClauses = append(setClauses, k+"=$"+string(rune('0'+i)))
		args = append(args, v)
		i++
	}
	setClauses = append(setClauses, "updated_at=NOW()")
	args = append(args, id)
	query := "UPDATE promo_codes SET " + join(setClauses, ", ") + " WHERE id=$" + string(rune('0'+i))
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *PromoCodeRepo) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM promo_codes WHERE id = $1", id)
	return err
}

func (r *PromoCodeRepo) GetByID(id uuid.UUID) (*PromoCode, error) {
	var promo PromoCode
	err := r.db.Get(&promo, "SELECT * FROM promo_codes WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &promo, nil
}

func (r *PromoCodeRepo) UpdatePromo(promo *PromoCode) error {
	_, err := r.db.NamedExec(`
		UPDATE promo_codes SET code=:code, discount_percent=:discount_percent, discount_fixed=:discount_fixed,
			usage_limit=:usage_limit, valid_until=:valid_until, is_active=:is_active, updated_at=NOW()
		WHERE id=:id
	`, promo)
	return err
}

func join(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
