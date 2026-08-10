DROP INDEX IF EXISTS idx_catalog_imports_publisher;
DROP INDEX IF EXISTS idx_catalog_imports_release_year;
DROP INDEX IF EXISTS idx_catalog_imports_game_section;
ALTER TABLE catalog_imports DROP COLUMN IF EXISTS publisher;
ALTER TABLE catalog_imports DROP COLUMN IF EXISTS release_year;
ALTER TABLE catalog_imports DROP COLUMN IF EXISTS game_section;
