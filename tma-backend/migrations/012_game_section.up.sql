ALTER TABLE products ADD COLUMN IF NOT EXISTS game_section VARCHAR(20) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_products_game_section ON products(game_section);
