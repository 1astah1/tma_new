-- Add new order statuses for invalid credentials and invalid 2FA
ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'CREDENTIALS_INVALID';
ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'INVALID_2FA';
