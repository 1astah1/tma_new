package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"tma-backend/internal/domain"
)

type CatalogImportRepo struct {
	db *sqlx.DB
}

func NewCatalogImportRepo(db *sqlx.DB) *CatalogImportRepo {
	return &CatalogImportRepo{db: db}
}

type CatalogImportFilter struct {
	Search      string
	Source      string
	Platform    string
	Status      string
	GameSection string
	ReleaseYear int
	Publisher   string
	Page        int
	Limit       int
}

type CatalogImportPublisherOption struct {
	Name  string `db:"publisher" json:"name"`
	Count int    `db:"count" json:"count"`
}

type CatalogImportYearOption struct {
	Year  int `db:"release_year" json:"year"`
	Count int `db:"count" json:"count"`
}

type CatalogImportFilterOptions struct {
	Publishers   []CatalogImportPublisherOption `json:"publishers"`
	ReleaseYears []CatalogImportYearOption      `json:"release_years"`
	Backfilled   bool                           `json:"backfilled"`
}

func (r *CatalogImportRepo) EnsureSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
ALTER TYPE platform_type ADD VALUE IF NOT EXISTS 'pc';

CREATE TABLE IF NOT EXISTS catalog_imports (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id        TEXT NOT NULL,
    source             TEXT NOT NULL,
    title              TEXT NOT NULL,
    description        TEXT,
    image_url          TEXT,
    platforms          TEXT[] NOT NULL DEFAULT '{}',
    original_price_rub DECIMAL(10,2),
    original_currency  VARCHAR(10),
    raw                JSONB NOT NULL DEFAULT '{}'::jsonb,
    status             TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    product_id         UUID REFERENCES products(id) ON DELETE SET NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source, external_id)
);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_status ON catalog_imports(status);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_source ON catalog_imports(source);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_platforms ON catalog_imports USING GIN(platforms);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_search ON catalog_imports USING GIN(to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(description, '')));
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS game_section VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS release_year INT;
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS publisher TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS release_date TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_catalog_imports_game_section ON catalog_imports(game_section);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_release_year ON catalog_imports(release_year);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_publisher ON catalog_imports(publisher);
ALTER TABLE products ADD COLUMN IF NOT EXISTS release_date TIMESTAMPTZ;
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS title_key TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS platform_family TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_catalog_imports_title_dedup ON catalog_imports(title_key, platform_family);
ALTER TABLE products ADD COLUMN IF NOT EXISTS title_key TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_products_title_dedup ON products(title_key, platform);
ALTER TABLE products ADD COLUMN IF NOT EXISTS vitrina_score DOUBLE PRECISION NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_products_vitrina_score ON products(vitrina_score DESC) WHERE type = 'game' AND status = 'active';
ALTER TABLE products ADD COLUMN IF NOT EXISTS prices JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS prices JSONB NOT NULL DEFAULT '{}'::jsonb;
`)
	if err != nil {
		return err
	}
	return r.BackfillImportMetadata(ctx)
}

func (r *CatalogImportRepo) BackfillImportMetadata(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE catalog_imports SET
    publisher = trim(CASE
        WHEN source = 'ps_store' THEN coalesce(raw->>'provider_name', '')
        ELSE coalesce(raw#>>'{LocalizedProperties,0,PublisherName}', '')
    END),
    release_year = sub.release_year,
    release_date = sub.release_at,
    game_section = sub.game_section
FROM (
    SELECT
        id,
        release_at,
        CASE
            WHEN release_at IS NULL THEN NULL
            ELSE EXTRACT(YEAR FROM release_at)::int
        END AS release_year,
        CASE
            WHEN is_preorder OR (release_at IS NOT NULL AND release_at > NOW()) THEN 'preorder'
            WHEN release_at IS NOT NULL AND release_at >= NOW() - INTERVAL '90 days' AND release_at <= NOW() THEN 'new'
            ELSE 'game'
        END AS game_section
    FROM (
        SELECT
            id,
            CASE
                WHEN source = 'ps_store' AND coalesce(raw->>'release_date', '') <> ''
                    THEN (raw->>'release_date')::timestamptz
                WHEN source = 'xbox_store' AND coalesce(raw#>>'{MarketProperties,0,OriginalReleaseDate}', '') <> ''
                    THEN (raw#>>'{MarketProperties,0,OriginalReleaseDate}')::timestamptz
                WHEN source = 'xbox_store' AND coalesce(raw#>>'{DisplaySkuAvailabilities,0,Sku,MarketProperties,0,FirstAvailableDate}', '') <> ''
                    THEN (raw#>>'{DisplaySkuAvailabilities,0,Sku,MarketProperties,0,FirstAvailableDate}')::timestamptz
                ELSE NULL
            END AS release_at,
            CASE
                WHEN source = 'xbox_store' AND lower(coalesce(raw#>>'{DisplaySkuAvailabilities,0,Sku,Properties,IsPreOrder}', 'false')) = 'true'
                    THEN TRUE
                ELSE FALSE
            END AS is_preorder
        FROM catalog_imports
    ) parsed
) sub
WHERE catalog_imports.id = sub.id`)
	return err
}

// BackfillDescriptionsByTitleKey переносит описание и картинку между позициями
// одной игры: у PS-версии текст может отсутствовать, а у Xbox-версии быть, и
// наоборот. Карточка одна, поэтому пустых полей быть не должно.
func (r *CatalogImportRepo) BackfillDescriptionsByTitleKey(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE catalog_imports target SET
    description = COALESCE(NULLIF(target.description, ''), donor.description),
    image_url = COALESCE(NULLIF(target.image_url, ''), donor.image_url),
    updated_at = NOW()
FROM (
    SELECT DISTINCT ON (title_key)
        title_key,
        description,
        image_url
    FROM catalog_imports
    WHERE title_key <> ''
      AND description IS NOT NULL
      AND length(description) > 60
    ORDER BY title_key, length(description) DESC
) donor
WHERE target.title_key = donor.title_key
  AND (target.description IS NULL OR length(target.description) <= 60 OR target.image_url IS NULL)`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *CatalogImportRepo) CountNeedingMetadata(ctx context.Context, source string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
SELECT COUNT(*) FROM catalog_imports
WHERE ($1 = '' OR source = $1)
  AND (publisher = '' OR release_year IS NULL)`, source)
	return count, err
}

func (r *CatalogImportRepo) ListExternalIDsBySource(ctx context.Context, source string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10000
	}
	var ids []string
	err := r.db.SelectContext(ctx, &ids, `
SELECT external_id FROM catalog_imports
WHERE ($1 = '' OR source = $1)
ORDER BY (publisher = '' OR publisher IS NULL) DESC, updated_at DESC
LIMIT $2`, source, limit)
	if ids == nil {
		ids = []string{}
	}
	return ids, err
}

func (r *CatalogImportRepo) CountWithPublisher(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM catalog_imports WHERE publisher <> ''`)
	return count, err
}

func (r *CatalogImportRepo) UpdateImportMetadata(ctx context.Context, item *domain.CatalogImport) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE catalog_imports SET
    publisher = $3,
    release_year = $4,
    release_date = $5,
    game_section = $6,
    original_price_rub = COALESCE($8, original_price_rub),
    original_currency = COALESCE($9, original_currency),
    raw = $7::jsonb,
    prices = COALESCE($10::jsonb, prices),
    updated_at = NOW()
WHERE source = $1 AND external_id = $2`,
		item.Source, item.ExternalID, item.Publisher, item.ReleaseYear, item.ReleaseDate, item.GameSection, string(item.Raw), item.OriginalPriceRUB, item.OriginalCurrency, string(item.Prices),
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *CatalogImportRepo) GetFilterOptions(ctx context.Context, backfill bool) (*CatalogImportFilterOptions, error) {
	opts := &CatalogImportFilterOptions{
		Publishers:   []CatalogImportPublisherOption{},
		ReleaseYears: []CatalogImportYearOption{},
	}
	if backfill {
		if err := r.BackfillImportMetadata(ctx); err != nil {
			return nil, err
		}
		opts.Backfilled = true
	}
	if err := r.db.SelectContext(ctx, &opts.Publishers, `
SELECT publisher, COUNT(*)::int AS count
FROM catalog_imports
WHERE publisher <> ''
GROUP BY publisher
ORDER BY count DESC, publisher
LIMIT 120`); err != nil {
		return nil, err
	}
	if err := r.db.SelectContext(ctx, &opts.ReleaseYears, `
SELECT release_year, COUNT(*)::int AS count
FROM catalog_imports
WHERE release_year IS NOT NULL
GROUP BY release_year
ORDER BY release_year DESC
LIMIT 30`); err != nil {
		return nil, err
	}
	return opts, nil
}

func (r *CatalogImportRepo) List(ctx context.Context, f CatalogImportFilter) ([]domain.CatalogImport, int, error) {
	args := []interface{}{}
	where := []string{}
	argIdx := 1

	if f.Search != "" {
		where = append(where, fmt.Sprintf("(title ILIKE '%%' || $%d || '%%' OR description ILIKE '%%' || $%d || '%%' OR source ILIKE '%%' || $%d || '%%')", argIdx, argIdx, argIdx))
		args = append(args, f.Search)
		argIdx++
	}
	if f.Source != "" {
		where = append(where, fmt.Sprintf("source = $%d", argIdx))
		args = append(args, f.Source)
		argIdx++
	}
	if f.Platform != "" {
		where = append(where, fmt.Sprintf("$%d = ANY(platforms)", argIdx))
		args = append(args, f.Platform)
		argIdx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.GameSection != "" {
		where = append(where, fmt.Sprintf("game_section = $%d", argIdx))
		args = append(args, f.GameSection)
		argIdx++
	}
	if f.ReleaseYear > 0 {
		where = append(where, fmt.Sprintf("release_year = $%d", argIdx))
		args = append(args, f.ReleaseYear)
		argIdx++
	}
	if f.Publisher != "" {
		where = append(where, fmt.Sprintf("publisher = $%d", argIdx))
		args = append(args, f.Publisher)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM catalog_imports"+whereClause, args...); err != nil {
		return nil, 0, err
	}

	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 25
	}

	args = append(args, f.Limit, (f.Page-1)*f.Limit)
	query := "SELECT * FROM catalog_imports" + whereClause + fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)

	var imports []domain.CatalogImport
	if err := r.db.SelectContext(ctx, &imports, query, args...); err != nil {
		return nil, 0, err
	}
	if imports == nil {
		imports = []domain.CatalogImport{}
	}
	return imports, total, nil
}

func (r *CatalogImportRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CatalogImport, error) {
	var item domain.CatalogImport
	if err := r.db.GetContext(ctx, &item, "SELECT * FROM catalog_imports WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CatalogImportRepo) UpsertFresh(ctx context.Context, item *domain.CatalogImport) error {
	err := r.db.GetContext(ctx, item, `
INSERT INTO catalog_imports (external_id, source, title, title_key, platform_family, description, image_url, platforms, game_section, release_year, release_date, publisher, original_price_rub, original_currency, raw, prices, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, COALESCE(NULLIF($15, ''), '{}')::jsonb, COALESCE(NULLIF($16, ''), '{}')::jsonb, 'pending')
ON CONFLICT (source, external_id) DO UPDATE SET
    title = EXCLUDED.title,
    title_key = EXCLUDED.title_key,
    platform_family = EXCLUDED.platform_family,
    description = EXCLUDED.description,
    image_url = EXCLUDED.image_url,
    platforms = EXCLUDED.platforms,
    original_price_rub = EXCLUDED.original_price_rub,
    original_currency = EXCLUDED.original_currency,
    game_section = EXCLUDED.game_section,
    release_year = EXCLUDED.release_year,
    release_date = EXCLUDED.release_date,
    publisher = EXCLUDED.publisher,
    raw = EXCLUDED.raw,
    prices = EXCLUDED.prices,
    status = 'pending',
    product_id = NULL,
    updated_at = NOW()
RETURNING *`,
		item.ExternalID, item.Source, item.Title, item.TitleKey, item.PlatformFamily, item.Description, item.ImageURL, item.Platforms,
		item.GameSection, item.ReleaseYear, item.ReleaseDate, item.Publisher,
		item.OriginalPriceRUB, item.OriginalCurrency, string(item.Raw), string(item.Prices),
	)
	return err
}

// UpsertWanted — upsert для адресного импорта по списку. В отличие от
// UpsertPending всегда перезаписывает название, описание и картинку: эти поля
// мы формируем сами из списка, и повторный прогон обязан их исправлять.
func (r *CatalogImportRepo) UpsertWanted(ctx context.Context, item *domain.CatalogImport) error {
	err := r.db.GetContext(ctx, item, `
INSERT INTO catalog_imports (external_id, source, title, title_key, platform_family, description, image_url, platforms, game_section, release_year, release_date, publisher, original_price_rub, original_currency, raw, prices, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, COALESCE(NULLIF($15, ''), '{}')::jsonb, COALESCE(NULLIF($16, ''), '{}')::jsonb, 'pending')
ON CONFLICT (source, external_id) DO UPDATE SET
    title = EXCLUDED.title,
    title_key = EXCLUDED.title_key,
    platform_family = EXCLUDED.platform_family,
    description = COALESCE(NULLIF(EXCLUDED.description, ''), catalog_imports.description),
    image_url = COALESCE(NULLIF(EXCLUDED.image_url, ''), catalog_imports.image_url),
    platforms = EXCLUDED.platforms,
    original_price_rub = EXCLUDED.original_price_rub,
    original_currency = EXCLUDED.original_currency,
    game_section = EXCLUDED.game_section,
    release_year = EXCLUDED.release_year,
    release_date = EXCLUDED.release_date,
    publisher = EXCLUDED.publisher,
    raw = EXCLUDED.raw,
    -- Цены сливаем, а не заменяем: если стор временно недоступен по одному
    -- региону, прогон вернёт только часть цен, и замена стёрла бы остальные.
    prices = CASE WHEN EXCLUDED.prices = '{}'::jsonb
                  THEN catalog_imports.prices
                  ELSE catalog_imports.prices || EXCLUDED.prices END,
    updated_at = NOW()
RETURNING *`,
		item.ExternalID, item.Source, item.Title, item.TitleKey, item.PlatformFamily, item.Description, item.ImageURL, item.Platforms,
		item.GameSection, item.ReleaseYear, item.ReleaseDate, item.Publisher,
		item.OriginalPriceRUB, item.OriginalCurrency, string(item.Raw), string(item.Prices),
	)
	return err
}

func (r *CatalogImportRepo) UpsertPending(ctx context.Context, item *domain.CatalogImport) error {
	err := r.db.GetContext(ctx, item, `
INSERT INTO catalog_imports (external_id, source, title, title_key, platform_family, description, image_url, platforms, game_section, release_year, release_date, publisher, original_price_rub, original_currency, raw, prices, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, COALESCE(NULLIF($15, ''), '{}')::jsonb, COALESCE(NULLIF($16, ''), '{}')::jsonb, 'pending')
ON CONFLICT (source, external_id) DO UPDATE SET
    title = CASE WHEN catalog_imports.status = 'pending' THEN EXCLUDED.title ELSE catalog_imports.title END,
    title_key = EXCLUDED.title_key,
    platform_family = EXCLUDED.platform_family,
    description = CASE WHEN catalog_imports.status = 'pending' THEN EXCLUDED.description ELSE catalog_imports.description END,
    image_url = CASE WHEN catalog_imports.status = 'pending' THEN EXCLUDED.image_url ELSE catalog_imports.image_url END,
    platforms = CASE WHEN catalog_imports.status = 'pending' THEN EXCLUDED.platforms ELSE catalog_imports.platforms END,
    original_price_rub = CASE WHEN catalog_imports.status = 'pending' THEN EXCLUDED.original_price_rub ELSE catalog_imports.original_price_rub END,
    original_currency = CASE WHEN catalog_imports.status = 'pending' THEN EXCLUDED.original_currency ELSE catalog_imports.original_currency END,
    game_section = EXCLUDED.game_section,
    release_year = EXCLUDED.release_year,
    release_date = EXCLUDED.release_date,
    publisher = EXCLUDED.publisher,
    raw = EXCLUDED.raw,
    prices = CASE WHEN EXCLUDED.prices = '{}'::jsonb THEN catalog_imports.prices ELSE EXCLUDED.prices END,
    updated_at = NOW()
RETURNING *`,
		item.ExternalID, item.Source, item.Title, item.TitleKey, item.PlatformFamily, item.Description, item.ImageURL, item.Platforms,
		item.GameSection, item.ReleaseYear, item.ReleaseDate, item.Publisher,
		item.OriginalPriceRUB, item.OriginalCurrency, string(item.Raw), string(item.Prices),
	)
	return err
}

func (r *CatalogImportRepo) MarkApproved(ctx context.Context, id, productID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE catalog_imports SET status='approved', product_id=$2, updated_at=NOW() WHERE id=$1", id, productID)
	return err
}

func (r *CatalogImportRepo) MarkRejected(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE catalog_imports SET status='rejected', updated_at=NOW() WHERE id=$1", id)
	return err
}

func (r *CatalogImportRepo) ResetRejectedImports(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE catalog_imports SET status='pending', updated_at=NOW() WHERE status='rejected'")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

type CatalogResetResult struct {
	DeletedImports      int64 `json:"deleted_imports"`
	DeletedProducts     int64 `json:"deleted_products"`
	DeactivatedProducts int64 `json:"deactivated_products"`
	DeletedKeys         int64 `json:"deleted_keys"`
}

func (r *CatalogImportRepo) ResetGameCatalog(ctx context.Context) (*CatalogResetResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result := &CatalogResetResult{}

	res, err := tx.ExecContext(ctx, "DELETE FROM catalog_imports")
	if err != nil {
		return nil, err
	}
	result.DeletedImports, _ = res.RowsAffected()

	res, err = tx.ExecContext(ctx, "DELETE FROM product_keys WHERE product_id IN (SELECT id FROM products WHERE type = 'game')")
	if err != nil {
		return nil, err
	}
	result.DeletedKeys, _ = res.RowsAffected()

	res, err = tx.ExecContext(ctx, `
DELETE FROM products p
WHERE p.type = 'game'
  AND NOT EXISTS (SELECT 1 FROM orders o WHERE o.product_id = p.id)
`)
	if err != nil {
		return nil, err
	}
	result.DeletedProducts, _ = res.RowsAffected()

	res, err = tx.ExecContext(ctx, `
UPDATE products p
SET status = 'inactive', updated_at = NOW()
WHERE p.type = 'game'
  AND EXISTS (SELECT 1 FROM orders o WHERE o.product_id = p.id)
`)
	if err != nil {
		return nil, err
	}
	result.DeactivatedProducts, _ = res.RowsAffected()

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

type CatalogDedupeResult struct {
	ImportsRejected int64 `json:"imports_rejected"`
	ProductsDeleted int64 `json:"products_deleted"`
	ProductsHidden  int64 `json:"products_hidden"`
}

func (r *CatalogImportRepo) UpdateImportKeys(ctx context.Context, id uuid.UUID, titleKey, platformFamily string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE catalog_imports SET title_key = $2, platform_family = $3, updated_at = NOW()
WHERE id = $1`, id, titleKey, platformFamily)
	return err
}

func (r *CatalogImportRepo) RemoveDuplicateImports(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
WITH ranked AS (
    SELECT
        id,
        status,
        ROW_NUMBER() OVER (
            PARTITION BY title_key, platform_family
            ORDER BY
                CASE status WHEN 'approved' THEN 0 WHEN 'pending' THEN 1 ELSE 2 END,
                CASE WHEN coalesce(image_url, '') <> '' THEN 0 ELSE 1 END,
                CASE WHEN publisher <> '' THEN 0 ELSE 1 END,
                CASE WHEN original_price_rub IS NOT NULL AND original_price_rub > 0 THEN 0 ELSE 1 END,
                updated_at DESC
        ) AS rn
    FROM catalog_imports
    WHERE title_key <> '' AND platform_family <> ''
)
UPDATE catalog_imports ci
SET status = 'rejected', updated_at = NOW()
FROM ranked r
WHERE ci.id = r.id
  AND r.rn > 1
  AND ci.status = 'pending'`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *CatalogImportRepo) ListPendingBySections(ctx context.Context, sections []string, limit int) ([]domain.CatalogImport, error) {
	if limit <= 0 {
		limit = 200
	}
	if len(sections) == 0 {
		return []domain.CatalogImport{}, nil
	}
	query, args, err := sqlx.In(`
SELECT * FROM catalog_imports
WHERE status = 'pending'
  AND game_section IN (?)
  AND title_key <> ''
ORDER BY
  CASE game_section WHEN 'preorder' THEN 0 WHEN 'new' THEN 1 ELSE 2 END,
  release_date NULLS LAST,
  updated_at DESC
LIMIT ?`, sections, limit)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var items []domain.CatalogImport
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.CatalogImport{}
	}
	return items, nil
}

func (r *CatalogImportRepo) FindApprovedByTitleKey(ctx context.Context, titleKey, platformFamily string) (*domain.CatalogImport, error) {
	if titleKey == "" {
		return nil, sql.ErrNoRows
	}
	var item domain.CatalogImport
	err := r.db.GetContext(ctx, &item, `
SELECT * FROM catalog_imports
WHERE title_key = $1 AND platform_family = $2 AND status = 'approved' AND product_id IS NOT NULL
ORDER BY updated_at DESC
LIMIT 1`, titleKey, platformFamily)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

type ImportKeyRow struct {
	ID        uuid.UUID      `db:"id"`
	Title     string         `db:"title"`
	Source    string         `db:"source"`
	Platforms pq.StringArray `db:"platforms"`
}

func (r *CatalogImportRepo) ListImportKeyRows(ctx context.Context) ([]ImportKeyRow, error) {
	var rows []ImportKeyRow
	err := r.db.SelectContext(ctx, &rows, `SELECT id, title, source, platforms FROM catalog_imports`)
	if rows == nil {
		rows = []ImportKeyRow{}
	}
	return rows, err
}

func (r *CatalogImportRepo) ListExternalIDsForEnrichment(ctx context.Context, limit int) ([]string, []string, error) {
	if limit <= 0 {
		limit = 2500
	}
	type row struct {
		Source     string `db:"source"`
		ExternalID string `db:"external_id"`
	}
	var rows []row
	err := r.db.SelectContext(ctx, &rows, `
SELECT source, external_id
FROM catalog_imports
WHERE product_id IS NOT NULL
   OR status = 'pending'
   OR original_price_rub IS NULL
   OR original_price_rub < 149
GROUP BY source, external_id
LIMIT $1`, limit)
	if err != nil {
		return nil, nil, err
	}
	psIDs := []string{}
	xboxIDs := []string{}
	seenPS := map[string]bool{}
	seenXbox := map[string]bool{}
	for _, item := range rows {
		switch item.Source {
		case domain.CatalogSourcePSStore:
			if !seenPS[item.ExternalID] {
				seenPS[item.ExternalID] = true
				psIDs = append(psIDs, item.ExternalID)
			}
		case domain.CatalogSourceXboxStore:
			if !seenXbox[item.ExternalID] {
				seenXbox[item.ExternalID] = true
				xboxIDs = append(xboxIDs, item.ExternalID)
			}
		}
	}
	return psIDs, xboxIDs, nil
}

func (r *CatalogImportRepo) ListAllPending(ctx context.Context, offset, limit int) ([]domain.CatalogImport, error) {
	if limit <= 0 {
		limit = 500
	}
	var items []domain.CatalogImport
	err := r.db.SelectContext(ctx, &items, `
SELECT * FROM catalog_imports
WHERE status = 'pending' AND title_key <> ''
ORDER BY
  CASE game_section WHEN 'preorder' THEN 0 WHEN 'new' THEN 1 ELSE 2 END,
  release_date NULLS LAST,
  updated_at DESC
OFFSET $1 LIMIT $2`, offset, limit)
	if items == nil {
		items = []domain.CatalogImport{}
	}
	return items, err
}

func (r *CatalogImportRepo) RequeueRejectedPreorders(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE catalog_imports
SET status = 'pending', updated_at = NOW()
WHERE status = 'rejected'
  AND game_section = 'preorder'
  AND release_date IS NOT NULL
  AND release_date > NOW()
  AND EXTRACT(YEAR FROM release_date) <= 2035`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *CatalogImportRepo) CountPending(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM catalog_imports WHERE status = 'pending'`)
	return count, err
}

func (r *CatalogImportRepo) ListPendingForQuality(ctx context.Context, limit int) ([]domain.CatalogImport, error) {
	if limit <= 0 {
		limit = 5000
	}
	var items []domain.CatalogImport
	err := r.db.SelectContext(ctx, &items, `
SELECT * FROM catalog_imports
WHERE status = 'pending'
ORDER BY updated_at DESC
LIMIT $1`, limit)
	if items == nil {
		items = []domain.CatalogImport{}
	}
	return items, err
}

func (r *CatalogImportRepo) ClearInvalidReleaseDates(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE catalog_imports
SET release_date = NULL, updated_at = NOW()
WHERE release_date IS NOT NULL
  AND (EXTRACT(YEAR FROM release_date) > 2035 OR EXTRACT(YEAR FROM release_date) < 1990)`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

type DescriptionRefreshRow struct {
	ImportID    uuid.UUID `db:"import_id"`
	ProductID   uuid.UUID `db:"product_id"`
	ExternalID  string    `db:"external_id"`
	Source      string    `db:"source"`
	Description string    `db:"description"`
}

func (r *CatalogImportRepo) ListLinkedForDescriptionRefresh(ctx context.Context, limit, offset int) ([]DescriptionRefreshRow, error) {
	if limit <= 0 {
		limit = 200
	}
	var rows []DescriptionRefreshRow
	err := r.db.SelectContext(ctx, &rows, `
SELECT ci.id AS import_id, p.id AS product_id, ci.external_id, ci.source, COALESCE(ci.description, '') AS description
FROM products p
JOIN catalog_imports ci ON ci.product_id = p.id AND ci.status = 'approved'
WHERE p.status = 'active' AND p.type = 'game'
ORDER BY p.updated_at ASC
LIMIT $1 OFFSET $2`, limit, offset)
	if rows == nil {
		rows = []DescriptionRefreshRow{}
	}
	return rows, err
}

func (r *CatalogImportRepo) UpdateDescription(ctx context.Context, importID uuid.UUID, description string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE catalog_imports SET description=$2, updated_at=NOW() WHERE id=$1`, importID, description)
	return err
}

func (r *CatalogImportRepo) GetByProductID(ctx context.Context, productID uuid.UUID) (*domain.CatalogImport, error) {
	var item domain.CatalogImport
	err := r.db.GetContext(ctx, &item, `
SELECT * FROM catalog_imports
WHERE product_id = $1
ORDER BY updated_at DESC
LIMIT 1`, productID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CatalogImportRepo) ListLinkedForSync(ctx context.Context, limit int) ([]domain.CatalogImport, error) {
	if limit <= 0 {
		limit = 5000
	}
	var items []domain.CatalogImport
	err := r.db.SelectContext(ctx, &items, `
SELECT * FROM catalog_imports
WHERE product_id IS NOT NULL AND status = 'approved'
ORDER BY updated_at DESC
LIMIT $1`, limit)
	if items == nil {
		items = []domain.CatalogImport{}
	}
	return items, err
}

type CatalogSummary struct {
	TotalImports         int `db:"total_imports" json:"total_imports"`
	PendingImports       int `db:"pending_imports" json:"pending_imports"`
	ApprovedImports      int `db:"approved_imports" json:"approved_imports"`
	RejectedImports      int `db:"rejected_imports" json:"rejected_imports"`
	GameProducts         int `db:"game_products" json:"game_products"`
	ActiveGameProducts   int `db:"active_game_products" json:"active_game_products"`
	InactiveGameProducts int `db:"inactive_game_products" json:"inactive_game_products"`
	OrphanGameProducts   int `db:"orphan_game_products" json:"orphan_game_products"`
}

func (r *CatalogImportRepo) Summary(ctx context.Context) (*CatalogSummary, error) {
	var summary CatalogSummary
	err := r.db.GetContext(ctx, &summary, `
SELECT
  (SELECT COUNT(*)::int FROM catalog_imports) AS total_imports,
  (SELECT COUNT(*)::int FROM catalog_imports WHERE status = 'pending') AS pending_imports,
  (SELECT COUNT(*)::int FROM catalog_imports WHERE status = 'approved') AS approved_imports,
  (SELECT COUNT(*)::int FROM catalog_imports WHERE status = 'rejected') AS rejected_imports,
  (SELECT COUNT(*)::int FROM products WHERE type = 'game') AS game_products,
  (SELECT COUNT(*)::int FROM products WHERE type = 'game' AND status = 'active') AS active_game_products,
  (SELECT COUNT(*)::int FROM products WHERE type = 'game' AND status = 'inactive') AS inactive_game_products,
  (SELECT COUNT(*)::int FROM products p WHERE p.type = 'game' AND NOT EXISTS (
    SELECT 1 FROM catalog_imports ci WHERE ci.product_id = p.id
  )) AS orphan_game_products`)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}
