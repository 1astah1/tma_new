ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS title_key TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS platform_family TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_catalog_imports_title_dedup ON catalog_imports(title_key, platform_family);

ALTER TABLE products ADD COLUMN IF NOT EXISTS title_key TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_products_title_dedup ON products(title_key, platform);
