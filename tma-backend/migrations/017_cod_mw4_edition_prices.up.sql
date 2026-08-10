-- Edition pricing for Call of Duty: Modern Warfare 4 preorder
UPDATE products
SET
  title = 'Call of Duty®: Modern Warfare® 4 (PS/XBOX/PC)',
  price = 5000,
  discount_percent = 0,
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
WHERE id = '63c124f6-af26-4cac-b5da-3be19905398a';
