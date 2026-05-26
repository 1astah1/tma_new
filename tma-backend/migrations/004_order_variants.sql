-- Add variant_id and quantity to orders
ALTER TABLE orders ADD COLUMN IF NOT EXISTS variant_id TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS quantity INT DEFAULT 1;
