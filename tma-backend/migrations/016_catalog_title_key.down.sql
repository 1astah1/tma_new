DROP INDEX IF EXISTS idx_products_title_dedup;
ALTER TABLE products DROP COLUMN IF EXISTS title_key;

DROP INDEX IF EXISTS idx_catalog_imports_title_dedup;
ALTER TABLE catalog_imports DROP COLUMN IF EXISTS platform_family;
ALTER TABLE catalog_imports DROP COLUMN IF EXISTS title_key;
