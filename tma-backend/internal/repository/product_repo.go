package repository

import (
	"context"
	"database/sql"
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
	Section  *string  `json:"section,omitempty"`
	Search   *string  `json:"search,omitempty"`
	MinPrice *float64 `json:"min_price,omitempty"`
	MaxPrice *float64 `json:"max_price,omitempty"`
	Status   *string  `json:"status,omitempty"`
	Sort     string   `json:"sort"`
	Order    string   `json:"order"`
	Page     int      `json:"page"`
	Limit    int      `json:"limit"`
}

// productGroupExpr — по чему считается «одна игра»: общий title_key, а если
// его нет, строка остаётся сама по себе.
const productGroupExpr = `COALESCE(NULLIF(title_key, ''), id::text)`

// productPlatformRankExpr — какую платформу показывать представителем карточки.
const productPlatformRankExpr = `CASE platform
    WHEN 'ps5' THEN 1
    WHEN 'xbox' THEN 2
    WHEN 'ps4' THEN 3
    WHEN 'pc' THEN 4
    ELSE 5 END`

const productColumns = `id, title, title_key, description, platform, type, game_section,
    release_date, price, discount_percent, variants, image_url, delivery_methods,
    status, vitrina_score, prices, created_at, updated_at`

func (r *ProductRepo) List(ctx context.Context, f ProductFilter) ([]domain.Product, int, error) {
	args := []interface{}{}
	where := []string{}

	argIdx := 1
	if f.Platform != nil && *f.Platform != "" {
		if *f.Platform == "ps" {
			where = append(where, "platform IN ('ps4', 'ps5')")
		} else {
			where = append(where, fmt.Sprintf("platform = $%d", argIdx))
			args = append(args, *f.Platform)
			argIdx++
		}
	}
	if f.Section != nil && *f.Section != "" {
		switch *f.Section {
		case "currency":
			where = append(where, "type = 'currency'")
		case "new":
			where = append(where, "type = 'game'", "game_section = 'new'")
		case "preorder":
			where = append(where, "type = 'game'", "game_section = 'preorder'")
		default:
			where = append(where, "type = 'game'")
		}
	} else if f.Type != nil && *f.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *f.Type)
		argIdx++
	}
	searchArg := 0
	if f.Search != nil && strings.TrimSpace(*f.Search) != "" {
		// Точное вхождение, все слова запроса в любом порядке, либо похожее
		// написание — иначе «call of dute» не находит Call of Duty.
		where = append(where, fmt.Sprintf(`(
        lower(title) LIKE '%%' || lower($%d) || '%%'
     OR title_key LIKE '%%' || lower($%d) || '%%'
     OR lower(title) %% lower($%d)
     OR title_key %% lower($%d)
)`, argIdx, argIdx, argIdx, argIdx))
		args = append(args, strings.TrimSpace(*f.Search))
		searchArg = argIdx
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

	// Считаем игры, а не строки: у одной игры отдельная строка на каждую
	// платформу, и без этого пагинация выдаёт одну и ту же игру дважды.
	countQuery := "SELECT COUNT(DISTINCT " + productGroupExpr + ") FROM products" + whereClause
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

	// Sorting
	sortField := "vitrina_score"
	sortOrder := "DESC"
	allowedSort := map[string]string{
		"price": "price", "title": "title", "created_at": "created_at", "release_date": "release_date", "vitrina_score": "vitrina_score",
	}
	if v, ok := allowedSort[f.Sort]; ok {
		sortField = v
	}
	if f.Order == "asc" {
		sortOrder = "ASC"
	}

	// На страницу выдаём по одной строке на игру — представителем берём
	// платформу с наибольшим приоритетом (PS5 → Xbox → PS4 → PC).
	//
	// В сортировке обязателен добивочный ключ id: у сотен игр одинаковый
	// vitrina_score, и без него Postgres волен возвращать строки в разном
	// порядке на каждой странице — одни игры показывались дважды, другие
	// не показывались вовсе.
	orderExpr := fmt.Sprintf("%s %s, id", sortField, sortOrder)
	if searchArg > 0 {
		orderExpr = fmt.Sprintf(
			"GREATEST(similarity(lower(title), lower($%d)), similarity(title_key, lower($%d))) DESC, %s %s, id",
			searchArg, searchArg, sortField, sortOrder)
	}

	query := fmt.Sprintf(`
SELECT %s FROM (
    SELECT DISTINCT ON (%s) *
    FROM products
    %s
    ORDER BY %s, %s, price ASC
) grouped
ORDER BY %s
LIMIT $%d OFFSET $%d`,
		productColumns, productGroupExpr, whereClause, productGroupExpr, productPlatformRankExpr,
		orderExpr, argIdx, argIdx+1)
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

func (r *ProductRepo) ListHomeFeedSectionCandidates(ctx context.Context, gameSection string, minPrice *float64, pool int, releaseNewestFirst bool) ([]domain.Product, error) {
	if pool <= 0 {
		pool = 80
	}
	order := "ASC"
	if releaseNewestFirst {
		order = "DESC"
	}

	args := []interface{}{gameSection}
	priceClause := ""
	if minPrice != nil {
		priceClause = " AND price >= $2"
		args = append(args, *minPrice)
	}

	shooterLimit := 60
	if pool < shooterLimit {
		shooterLimit = pool
	}

	shooterSQL := fmt.Sprintf(`
SELECT * FROM products
WHERE type = 'game' AND status = 'active' AND game_section = $1%s
  AND (
    title_key ILIKE '%%call of duty%%' OR title_key ILIKE '%%battlefield%%' OR
    title_key ILIKE '%%counter-strike%%' OR title_key ILIKE '%%helldivers%%' OR
    title_key ILIKE '%%apex legends%%' OR title_key ILIKE '%%overwatch%%' OR
    title_key ILIKE '%%arc raiders%%' OR title_key ILIKE '%%destiny%%' OR
    title_key ILIKE '%%rainbow six%%' OR title_key ILIKE '%%doom%%' OR
    title_key ILIKE '%%halo%%' OR title_key ILIKE '%%fortnite%%' OR
    title_key ILIKE '%%the finals%%' OR title_key ILIKE '%%pubg%%' OR
    title_key ILIKE '%%warzone%%' OR title_key ILIKE '%%far cry%%' OR
    title_key ILIKE '%%borderlands%%' OR title_key ILIKE '%%titanfall%%' OR
    title_key ILIKE '%%hell let loose%%' OR title_key ILIKE '%%sniper elite%%'
  )
ORDER BY CASE WHEN title_key ILIKE '%%call of duty%%' THEN 0 ELSE 1 END, release_date %s NULLS LAST
LIMIT %d`, priceClause, order, shooterLimit)

	var shooters []domain.Product
	if err := r.db.SelectContext(ctx, &shooters, shooterSQL, args...); err != nil {
		return nil, err
	}

	regularSQL := fmt.Sprintf(`
SELECT * FROM products
WHERE type = 'game' AND status = 'active' AND game_section = $1%s
ORDER BY release_date %s NULLS LAST
LIMIT %d`, priceClause, order, pool)

	var regular []domain.Product
	if err := r.db.SelectContext(ctx, &regular, regularSQL, args...); err != nil {
		return nil, err
	}

	return mergeProductsByID(shooters, regular), nil
}

func mergeProductsByID(parts ...[]domain.Product) []domain.Product {
	seen := map[uuid.UUID]bool{}
	out := make([]domain.Product, 0)
	for _, list := range parts {
		for _, p := range list {
			if seen[p.ID] {
				continue
			}
			seen[p.ID] = true
			out = append(out, p)
		}
	}
	return out
}

func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var p domain.Product
	err := r.db.GetContext(ctx, &p, "SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Product, error) {
	if len(ids) == 0 {
		return []domain.Product{}, nil
	}
	args := make([]interface{}, 0, len(ids))
	orderParts := make([]string, 0, len(ids))
	placeholders := make([]string, 0, len(ids))
	for i, id := range ids {
		arg := i + 1
		args = append(args, id)
		placeholders = append(placeholders, fmt.Sprintf("$%d", arg))
		orderParts = append(orderParts, fmt.Sprintf("WHEN $%d THEN %d", arg, i))
	}
	query := fmt.Sprintf(
		"SELECT * FROM products WHERE id IN (%s) ORDER BY CASE id %s END",
		strings.Join(placeholders, ","),
		strings.Join(orderParts, " "),
	)
	var products []domain.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	err := r.db.GetContext(ctx, p,
		`INSERT INTO products (title, title_key, description, platform, type, game_section, release_date, price, discount_percent, variants, image_url, delivery_methods, status, vitrina_score, prices)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 RETURNING *`, p.Title, p.TitleKey, p.Description, p.Platform, p.Type, p.GameSection, p.ReleaseDate, p.Price, p.DiscountPercent, p.Variants, p.ImageURL, p.DeliveryMethods, p.Status, p.VitrinaScore, p.Prices)
	return err
}

func (r *ProductRepo) UpdateVitrinaScore(ctx context.Context, id uuid.UUID, score float64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET vitrina_score = $2, updated_at = NOW() WHERE id = $1`, id, score)
	return err
}

func (r *ProductRepo) ListActiveGamesForScoring(ctx context.Context) ([]domain.Product, error) {
	var rows []domain.Product
	err := r.db.SelectContext(ctx, &rows, `SELECT * FROM products WHERE type = 'game' AND status = 'active'`)
	if rows == nil {
		rows = []domain.Product{}
	}
	return rows, err
}

func (r *ProductRepo) OrderCountsByTitleKey(ctx context.Context) (map[string]int, error) {
	type row struct {
		TitleKey string `db:"title_key"`
		Count    int    `db:"cnt"`
	}
	var rows []row
	err := r.db.SelectContext(ctx, &rows, `
		SELECT COALESCE(NULLIF(p.title_key, ''), LOWER(TRIM(p.title))) AS title_key, COUNT(*)::int AS cnt
		FROM orders o
		JOIN products p ON p.id = o.product_id
		WHERE p.type = 'game'
		  AND o.status NOT IN ('CANCELLED', 'REFUNDED', 'REFUND_REQUESTED')
		GROUP BY 1`)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(rows))
	for _, item := range rows {
		if item.TitleKey == "" {
			continue
		}
		out[item.TitleKey] = item.Count
	}
	return out, nil
}

func (r *ProductRepo) ListHybridPopular(ctx context.Context, pinnedOrder []uuid.UUID, limit int, minPrice float64) ([]domain.Product, error) {
	result := make([]domain.Product, 0, limit)
	seen := map[uuid.UUID]bool{}

	if len(pinnedOrder) > 0 {
		pinned, err := r.GetByIDs(ctx, pinnedOrder)
		if err != nil {
			return nil, err
		}
		for _, p := range pinned {
			if p.Status != domain.ProductStatusActive || p.Type != domain.ProductTypeGame {
				continue
			}
			if p.Price < minPrice && p.GameSection != "preorder" {
				continue
			}
			result = append(result, p)
			seen[p.ID] = true
			if len(result) >= limit {
				return result, nil
			}
		}
	}

	remaining := limit - len(result)
	if remaining <= 0 {
		return result, nil
	}

	existingTitleKeys := titleKeysFromProducts(result)
	fetchLimit := remaining * 3
	if fetchLimit < remaining+12 {
		fetchLimit = remaining + 12
	}

	var extra []domain.Product
	query := `
SELECT * FROM products
WHERE type = 'game' AND status = 'active' AND price >= $1`
	if len(seen) > 0 {
		ids := make([]uuid.UUID, 0, len(seen))
		for id := range seen {
			ids = append(ids, id)
		}
		q, a, err := sqlx.In(query+` AND id NOT IN (?) ORDER BY vitrina_score DESC, created_at DESC LIMIT ?`, ids, fetchLimit)
		if err != nil {
			return result, err
		}
		q = r.db.Rebind(q)
		err = r.db.SelectContext(ctx, &extra, q, a...)
		if err != nil {
			return result, err
		}
	} else {
		err := r.db.SelectContext(ctx, &extra, query+` ORDER BY vitrina_score DESC, created_at DESC LIMIT $2`, minPrice, fetchLimit)
		if err != nil {
			return result, err
		}
	}
	deduped := dedupeProductsByTitleKey(extra, existingTitleKeys)
	if len(deduped) > remaining {
		deduped = deduped[:remaining]
	}
	result = append(result, deduped...)
	return result, nil
}

func titleKeysFromProducts(products []domain.Product) map[string]bool {
	keys := make(map[string]bool, len(products))
	for _, p := range products {
		key := p.TitleKey
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(p.Title))
		}
		if key != "" {
			keys[key] = true
		}
	}
	return keys
}

func dedupeProductsByTitleKey(products []domain.Product, skip map[string]bool) []domain.Product {
	if len(products) == 0 {
		return products
	}
	seen := map[string]bool{}
	for key := range skip {
		seen[key] = true
	}
	out := make([]domain.Product, 0, len(products))
	for _, p := range products {
		key := p.TitleKey
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(p.Title))
		}
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE products SET title=:title, title_key=:title_key, description=:description, platform=:platform, 
		 type=:type, game_section=:game_section, release_date=:release_date, price=:price, discount_percent=:discount_percent, variants=:variants, image_url=:image_url, 
		 delivery_methods=:delivery_methods, status=:status, prices=:prices, updated_at=NOW()
		 WHERE id=:id`, p)
	return err
}

func (r *ProductRepo) UpdateDescription(ctx context.Context, id uuid.UUID, description string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET description=$2, updated_at=NOW() WHERE id=$1`, id, description)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE products SET status='inactive' WHERE id=$1", id)
	return err
}

func (r *ProductRepo) ActivateAllGames(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, "UPDATE products SET status='active', updated_at=NOW() WHERE type='game' AND status <> 'active'")
	if err != nil {
		return 0, err
	}
	count, _ := res.RowsAffected()
	return count, nil
}

func (r *ProductRepo) CountAvailableKeys(ctx context.Context, productID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM product_keys WHERE product_id = $1 AND status = 'available'", productID)
	return count, err
}

func (r *ProductRepo) CountOrders(ctx context.Context, productID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM orders WHERE product_id = $1", productID)
	return count, err
}

func (r *ProductRepo) UpdateTitleKey(ctx context.Context, id uuid.UUID, titleKey string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET title_key = $2, updated_at = NOW() WHERE id = $1`, id, titleKey)
	return err
}

func (r *ProductRepo) ListActiveByTitleKey(ctx context.Context, titleKey string) ([]domain.Product, error) {
	if titleKey == "" {
		return []domain.Product{}, nil
	}
	var products []domain.Product
	err := r.db.SelectContext(ctx, &products, `
SELECT * FROM products
WHERE type = 'game' AND status = 'active' AND title_key = $1
ORDER BY CASE platform
  WHEN 'ps5' THEN 0
  WHEN 'ps4' THEN 1
  WHEN 'xbox' THEN 2
  WHEN 'pc' THEN 3
  ELSE 4
END, vitrina_score DESC`, titleKey)
	if err != nil {
		return nil, err
	}
	if products == nil {
		products = []domain.Product{}
	}
	return products, nil
}

func (r *ProductRepo) FindByTitleKeyPlatform(ctx context.Context, titleKey string, platform domain.Platform) (*domain.Product, error) {
	if titleKey == "" {
		return nil, sql.ErrNoRows
	}
	var p domain.Product
	err := r.db.GetContext(ctx, &p, `
SELECT * FROM products
WHERE type = 'game' AND title_key = $1 AND platform = $2
ORDER BY CASE status WHEN 'active' THEN 0 ELSE 1 END, updated_at DESC
LIMIT 1`, titleKey, platform)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) RemoveDuplicateProducts(ctx context.Context) (deleted, hidden int64, err error) {
	res, err := r.db.ExecContext(ctx, `
WITH ranked AS (
    SELECT
        p.id,
        ROW_NUMBER() OVER (
            PARTITION BY p.title_key, p.platform
            ORDER BY
                CASE WHEN EXISTS (SELECT 1 FROM orders o WHERE o.product_id = p.id) THEN 0 ELSE 1 END,
                CASE p.status WHEN 'active' THEN 0 ELSE 1 END,
                p.updated_at DESC
        ) AS rn
    FROM products p
    WHERE p.type = 'game' AND p.title_key <> ''
)
DELETE FROM products p
USING ranked r
WHERE p.id = r.id
  AND r.rn > 1
  AND NOT EXISTS (SELECT 1 FROM orders o WHERE o.product_id = p.id)`)
	if err != nil {
		return 0, 0, err
	}
	deleted, _ = res.RowsAffected()

	res, err = r.db.ExecContext(ctx, `
WITH ranked AS (
    SELECT
        p.id,
        ROW_NUMBER() OVER (
            PARTITION BY p.title_key, p.platform
            ORDER BY
                CASE WHEN EXISTS (SELECT 1 FROM orders o WHERE o.product_id = p.id) THEN 0 ELSE 1 END,
                CASE p.status WHEN 'active' THEN 0 ELSE 1 END,
                p.updated_at DESC
        ) AS rn
    FROM products p
    WHERE p.type = 'game' AND p.title_key <> ''
)
UPDATE products p
SET status = 'inactive', updated_at = NOW()
FROM ranked r
WHERE p.id = r.id
  AND r.rn > 1
  AND EXISTS (SELECT 1 FROM orders o WHERE o.product_id = p.id)
  AND p.status = 'active'`)
	if err != nil {
		return deleted, 0, err
	}
	hidden, _ = res.RowsAffected()
	return deleted, hidden, nil
}

func (r *ProductRepo) ActivateFreshGames(ctx context.Context, minPrice float64) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE products SET status = 'active', updated_at = NOW()
WHERE type = 'game'
  AND game_section IN ('preorder', 'new')
  AND status <> 'active'
  AND (
    (game_section = 'preorder' AND release_date IS NOT NULL AND release_date > NOW() AND EXTRACT(YEAR FROM release_date) <= 2035)
    OR (price IS NOT NULL AND price >= $1)
  )`, minPrice)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *ProductRepo) ClearInvalidReleaseDates(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE products
SET release_date = NULL, updated_at = NOW()
WHERE release_date IS NOT NULL
  AND (EXTRACT(YEAR FROM release_date) > 2035 OR EXTRACT(YEAR FROM release_date) < 1990)`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *ProductRepo) StatusByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	query, args, err := sqlx.In(`SELECT id, status FROM products WHERE id IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	type row struct {
		ID     uuid.UUID `db:"id"`
		Status string    `db:"status"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.ID] = item.Status
	}
	return result, nil
}

// SyncCardFromImports подтягивает в карточку название, описание и картинку из
// связанного импорта: цены обновляет SyncMetadataFromImports, а текстовая часть
// без этого застывает на том, что было в момент создания карточки.
func (r *ProductRepo) ListGamesForTitleFix(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.SelectContext(ctx, &products,
		`SELECT * FROM products WHERE type = 'game' ORDER BY created_at`)
	return products, err
}

func (r *ProductRepo) UpdateTitle(ctx context.Context, id uuid.UUID, title, titleKey string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET title = $2, title_key = $3, updated_at = NOW() WHERE id = $1`,
		id, title, titleKey)
	return err
}

func (r *ProductRepo) SyncCardFromImports(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE products p
SET
    title = ci.title,
    title_key = ci.title_key,
    description = COALESCE(NULLIF(ci.description, ''), p.description),
    image_url = COALESCE(NULLIF(ci.image_url, ''), p.image_url),
    updated_at = NOW()
FROM catalog_imports ci
WHERE ci.product_id = p.id
  AND p.type = 'game'
  AND ci.title <> ''
  AND (p.title <> ci.title OR p.title_key <> ci.title_key
       OR (p.description IS NULL AND ci.description IS NOT NULL)
       OR (p.image_url IS NULL AND ci.image_url IS NOT NULL))`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *ProductRepo) SyncMetadataFromImports(ctx context.Context, minPrice float64) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE products p
SET
    price = ci.original_price_rub,
    prices = ci.prices,
    release_date = ci.release_date,
    game_section = ci.game_section,
    updated_at = NOW()
FROM catalog_imports ci
WHERE ci.product_id = p.id
  AND ci.original_price_rub IS NOT NULL
  AND ci.original_price_rub >= $1
  AND p.type = 'game'`, minPrice)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *ProductRepo) DeactivateUnsellableGames(ctx context.Context, minPrice float64) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE products
SET status = 'inactive', updated_at = NOW()
WHERE type = 'game'
  AND status = 'active'
  AND NOT (
    game_section = 'preorder'
    AND release_date IS NOT NULL
    AND release_date > NOW()
    AND EXTRACT(YEAR FROM release_date) <= 2035
  )
  AND (price IS NULL OR price < $1 OR price <= 1.01)`, minPrice)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *ProductRepo) ListActiveGameTitles(ctx context.Context) ([]struct {
	ID    uuid.UUID `db:"id"`
	Title string    `db:"title"`
}, error) {
	var rows []struct {
		ID    uuid.UUID `db:"id"`
		Title string    `db:"title"`
	}
	err := r.db.SelectContext(ctx, &rows, `SELECT id, title FROM products WHERE type = 'game' AND status = 'active'`)
	return rows, err
}

func (r *ProductRepo) DeactivateByIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In(`UPDATE products SET status = 'inactive', updated_at = NOW() WHERE id IN (?)`, ids)
	if err != nil {
		return 0, err
	}
	query = r.db.Rebind(query)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *ProductRepo) ListGameIDs(ctx context.Context) ([]struct {
	ID    uuid.UUID `db:"id"`
	Title string    `db:"title"`
}, error) {
	var rows []struct {
		ID    uuid.UUID `db:"id"`
		Title string    `db:"title"`
	}
	err := r.db.SelectContext(ctx, &rows, `SELECT id, title FROM products WHERE type = 'game'`)
	return rows, err
}

func (r *ProductRepo) GetImportLinkByProductID(ctx context.Context, productID uuid.UUID) (externalID string, source string, ok bool) {
	err := r.db.QueryRowContext(ctx, `
SELECT external_id, source
FROM catalog_imports
WHERE product_id = $1 AND external_id <> ''
ORDER BY updated_at DESC
LIMIT 1`, productID).Scan(&externalID, &source)
	if err != nil {
		return "", "", false
	}
	return externalID, source, true
}

// ProductMatchRow — карточка вместе с названием позиции в сторе, к которой
// она привязана. Нужна для проверки, что карточка ведёт на саму игру.
type ProductMatchRow struct {
	ID         uuid.UUID `db:"id"`
	Title      string    `db:"title"`
	Platform   string    `db:"platform"`
	StoreTitle string    `db:"store_title"`
	Source     string    `db:"source"`
	ImportID   uuid.UUID `db:"import_id"`
}

func (r *ProductRepo) ListMatchesForAudit(ctx context.Context) ([]ProductMatchRow, error) {
	var rows []ProductMatchRow
	err := r.db.SelectContext(ctx, &rows, `
SELECT p.id,
       p.title,
       p.platform::text AS platform,
       coalesce(ci.raw->>'name', ci.raw#>>'{LocalizedProperties,0,ProductTitle}', '') AS store_title,
       ci.source,
       ci.id AS import_id
FROM products p
JOIN catalog_imports ci ON ci.product_id = p.id
WHERE p.type = 'game' AND p.status = 'active'`)
	return rows, err
}

func (r *ProductRepo) DeactivateByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET status = 'inactive', updated_at = NOW() WHERE id = $1`, id)
	return err
}
