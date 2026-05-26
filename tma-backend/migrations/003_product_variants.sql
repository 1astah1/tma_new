-- Add discount and variants to products
ALTER TABLE products ADD COLUMN IF NOT EXISTS discount_percent DECIMAL(5,2) DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS variants JSONB DEFAULT '[]'::jsonb;

-- Update existing products to have empty variants array
UPDATE products SET variants = '[]'::jsonb WHERE variants IS NULL;
