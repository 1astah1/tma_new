INSERT INTO settings (key, value) VALUES
('payment_details', '{
  "sbp": {"phone": "89841157865", "bank": "Альфа-Банк", "receiver": "Олеся К."},
  "card": {"number": "2200153684839138", "bank": "Альфа-Банк"},
  "crypto": {
    "binance": "143915969",
    "bybit": "100543830",
    "trc20": "TCZxsXBe8S1BiSVPEpS12UzsaxQjkHmgap"
  }
}'::jsonb)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
