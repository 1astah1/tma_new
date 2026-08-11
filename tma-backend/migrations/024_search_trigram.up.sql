-- Поиск по названию с опечатками: без триграмм «call of dute» не находит
-- ничего, хотя человек ищет Call of Duty.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_products_title_trgm
    ON products USING GIN (lower(title) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_title_key_trgm
    ON products USING GIN (title_key gin_trgm_ops);
