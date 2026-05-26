-- Migration 008: Key management improvements + user ban + product key counts

-- 1. Add UNIQUE constraint on product_keys(product_id, key) to prevent duplicates
ALTER TABLE product_keys ADD CONSTRAINT uq_product_key UNIQUE (product_id, key);

-- 2. Add 'invalid' status to key_status enum
ALTER TYPE key_status ADD VALUE IF NOT EXISTS 'invalid';

-- 3. Add user ban fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_banned BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS admin_notes TEXT DEFAULT '';

-- 4. Add promo code management columns (if not exists)
ALTER TABLE promo_codes ADD COLUMN IF NOT EXISTS usage_limit INTEGER DEFAULT NULL;
ALTER TABLE promo_codes ADD COLUMN IF NOT EXISTS used_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE promo_codes ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
