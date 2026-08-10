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
