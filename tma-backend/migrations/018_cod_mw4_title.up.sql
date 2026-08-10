-- Rename Call of Duty MW4 display title
UPDATE products
SET title = 'Call of Duty®: Modern Warfare® 4 (PS/XBOX/PC)', updated_at = NOW()
WHERE id = '63c124f6-af26-4cac-b5da-3be19905398a';

UPDATE catalog_imports
SET title = 'Call of Duty®: Modern Warfare® 4 (PS/XBOX/PC)', updated_at = NOW()
WHERE product_id = '63c124f6-af26-4cac-b5da-3be19905398a';
