-- Unified edition pricing for all Call of Duty: Modern Warfare 4 SKUs (PS + Xbox)
UPDATE products
SET
  title = 'Call of Duty®: Modern Warfare® 4 (PS/XBOX/PC)',
  title_key = 'call of duty: modern warfare 4',
  prices = '{
    "xbox": 5000,
    "edition_catalog": {
      "ps_tr": [
        {"id": "standard", "name": "Standard Edition", "price": 6300},
        {"id": "vault", "name": "Vault Edition", "price": 7800, "discount_label": "−10%"}
      ],
      "ps_ua": [
        {"id": "standard", "name": "Standard Edition", "price": 7300},
        {"id": "vault", "name": "Vault Edition", "price": 8900, "discount_label": "−10%"}
      ],
      "xbox": [
        {"id": "standard", "name": "Standard Edition", "price": 5000},
        {"id": "vault", "name": "Vault Edition", "price": 6700, "discount_label": "−10%"}
      ]
    }
  }'::jsonb,
  updated_at = NOW()
WHERE title_key = 'call of duty: modern warfare 4';
