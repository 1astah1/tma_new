-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id       BIGINT UNIQUE NOT NULL,
    username          VARCHAR(255),
    first_name        VARCHAR(255),
    last_interaction  TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_telegram_id ON users(telegram_id);

CREATE TYPE platform_type AS ENUM ('ps4', 'ps5', 'xbox');
CREATE TYPE product_type AS ENUM ('game', 'currency', 'subscription');
CREATE TYPE product_status AS ENUM ('active', 'inactive');

CREATE TABLE products (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    platform          platform_type NOT NULL,
    type              product_type NOT NULL,
    price             DECIMAL(10,2) NOT NULL,
    image_url         TEXT,
    delivery_methods  TEXT[] NOT NULL DEFAULT '{}',
    status            product_status NOT NULL DEFAULT 'active',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_products_platform ON products(platform);
CREATE INDEX idx_products_type ON products(type);
CREATE INDEX idx_products_status ON products(status);

CREATE TYPE order_status AS ENUM (
    'NEW','WAITING_PAYMENT','PAYMENT_VERIFICATION','PAID',
    'WAITING_ACTIVATION','AWAITING_CREDENTIALS','CREDENTIALS_RECEIVED',
    'AWAITING_2FA','ACTIVATING','ACTIVATED','KEY_ISSUED','COMPLETED',
    'CANCELLED','REFUND_REQUESTED','REFUNDED'
);
CREATE TYPE delivery_method AS ENUM ('key', 'activation');
CREATE TYPE payment_method AS ENUM ('sbp', 'card', 'crypto', 'telegram_stars');

CREATE TABLE orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    product_id          UUID NOT NULL REFERENCES products(id),
    delivery_method     delivery_method NOT NULL,
    status              order_status NOT NULL DEFAULT 'NEW',
    payment_method      VARCHAR(50),
    payment_amount      DECIMAL(10,2),
    payment_receipt_url TEXT,
    payment_verified_by UUID,
    key_id              UUID,
    assigned_admin_id   UUID,
    cancelled_reason    TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_assigned_admin ON orders(assigned_admin_id);
CREATE INDEX idx_orders_created_at ON orders(created_at);

CREATE TYPE changed_by_type AS ENUM ('admin', 'system', 'user');

CREATE TABLE order_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders(id),
    old_status      order_status,
    new_status      order_status NOT NULL,
    changed_by_id   UUID,
    changed_by_type changed_by_type NOT NULL DEFAULT 'system',
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_history_order ON order_history(order_id);
CREATE INDEX idx_order_history_created ON order_history(created_at);

CREATE TYPE key_status AS ENUM ('available', 'reserved', 'sold');

CREATE TABLE product_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    status      key_status NOT NULL DEFAULT 'available',
    order_id    UUID REFERENCES orders(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_keys_product_id ON product_keys(product_id);
CREATE INDEX idx_keys_status ON product_keys(status);

CREATE TABLE user_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    order_id        UUID NOT NULL REFERENCES orders(id),
    platform        platform_type NOT NULL,
    login           TEXT NOT NULL,
    password        TEXT NOT NULL,
    two_factor_code TEXT,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_user_accounts_user ON user_accounts(user_id);
CREATE INDEX idx_user_accounts_order ON user_accounts(order_id);

CREATE TABLE admins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id   BIGINT UNIQUE NOT NULL,
    username      VARCHAR(255),
    password_hash TEXT,
    roles         TEXT[] NOT NULL DEFAULT '{}',
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_admins_telegram ON admins(telegram_id);

CREATE TABLE admin_actions_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID NOT NULL REFERENCES admins(id),
    action_type VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id   UUID,
    details     JSONB,
    ip_address  VARCHAR(45),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_admin_log_admin ON admin_actions_log(admin_id);
CREATE INDEX idx_admin_log_type ON admin_actions_log(action_type);
CREATE INDEX idx_admin_log_created ON admin_actions_log(created_at);

CREATE TYPE transaction_status AS ENUM ('pending', 'verified', 'failed');

CREATE TABLE payment_transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id),
    method      VARCHAR(50) NOT NULL,
    amount      DECIMAL(10,2) NOT NULL,
    currency    VARCHAR(10) NOT NULL DEFAULT 'RUB',
    receipt_url TEXT,
    status      transaction_status NOT NULL DEFAULT 'pending',
    verified_by UUID REFERENCES admins(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payment_transactions_order ON payment_transactions(order_id);

CREATE TABLE settings (
    id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key   VARCHAR(255) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO users (telegram_id, username, first_name) VALUES
    (123456789, 'test_user', 'Test'),
    (987654321, 'demo_user', 'Demo');

INSERT INTO products (title, description, platform, type, price, delivery_methods, status) VALUES
    ('Cyberpunk 2077', 'Open-world RPG in Night City', 'xbox', 'game', 2499.00, '{key,activation}', 'active'),
    ('FIFA 24', 'Football simulation game', 'ps5', 'game', 3999.00, '{key}', 'active'),
    ('God of War Ragnarok', 'Epic Norse adventure', 'ps4', 'game', 3499.00, '{key,activation}', 'active'),
    ('V-Bucks 5000', 'Fortnite in-game currency', 'xbox', 'currency', 999.00, '{key}', 'active'),
    ('Game Pass Ultimate', '1 month subscription', 'xbox', 'subscription', 1499.00, '{key,activation}', 'active'),
    ('PS Plus Extra', '3 months subscription', 'ps5', 'subscription', 3999.00, '{key}', 'active');

INSERT INTO product_keys (product_id, key, status) VALUES
    ((SELECT id FROM products WHERE title = 'Cyberpunk 2077'), 'XXXXX-XXXXX-XXXXX-XXXX', 'available'),
    ((SELECT id FROM products WHERE title = 'Cyberpunk 2077'), 'YYYYY-YYYYY-YYYYY-YYYY', 'available'),
    ((SELECT id FROM products WHERE title = 'FIFA 24'), 'ZZZZZ-ZZZZZ-ZZZZZ-ZZZZ', 'available'),
    ((SELECT id FROM products WHERE title = 'V-Bucks 5000'), 'AAAAA-AAAAA-AAAAA-AAAA', 'available');

INSERT INTO admins (telegram_id, username, password_hash, roles, is_active) VALUES
    (111111, 'superadmin', '$2a$10$dummyhashfordevonly1234567890abcdef', '{super_admin,game_manager,activation_admin,support,finance}', true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS admin_actions_log;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS user_accounts;
DROP TABLE IF EXISTS product_keys;
DROP TABLE IF EXISTS order_history;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS settings;
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS changed_by_type;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS delivery_method;
DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS key_status;
DROP TYPE IF EXISTS product_status;
DROP TYPE IF EXISTS product_type;
DROP TYPE IF EXISTS platform_type;
-- +goose StatementEnd
