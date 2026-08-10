ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS game_section VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS release_year INT;
ALTER TABLE catalog_imports ADD COLUMN IF NOT EXISTS publisher TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_catalog_imports_game_section ON catalog_imports(game_section);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_release_year ON catalog_imports(release_year);
CREATE INDEX IF NOT EXISTS idx_catalog_imports_publisher ON catalog_imports(publisher);
