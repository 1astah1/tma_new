-- Колонки, которые до этого создавались только в рантайме (CatalogImportRepo.EnsureSchema),
-- из-за чего свежая база после миграций не совпадала с моделью в коде.
ALTER TABLE products ADD COLUMN IF NOT EXISTS prices JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE products ADD COLUMN IF NOT EXISTS vitrina_score DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS title_key TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS release_date TIMESTAMPTZ;

ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS prices JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS title_key TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS platform_family TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS game_section VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS release_year INT;
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS publisher TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS release_date TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_products_vitrina_score
    ON products(vitrina_score DESC) WHERE type = 'game' AND status = 'active';
CREATE INDEX IF NOT EXISTS idx_products_title_dedup ON products(title_key, platform);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_title_dedup ON catalog_imports(title_key, platform_family);
