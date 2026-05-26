INSERT INTO settings (key, value) VALUES
('payment_details', '{
  "sbp": {"phone": "89841157865", "bank": "РђР»СЊС„Р°-Р‘Р°РЅРє", "receiver": "РћР»РµСЃСЏ Рљ."},
  "card": {"number": "2200153684839138", "bank": "РђР»СЊС„Р°-Р‘Р°РЅРє"},
  "crypto": {
    "binance": "143915969",
    "bybit": "100543830",
    "trc20": "TCZxsXBe8S1BiSVPEpS12UzsaxQjkHmgap"
  }
}'::jsonb)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
