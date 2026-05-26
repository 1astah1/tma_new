-- Order chat messages
CREATE TABLE IF NOT EXISTS order_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sender_type VARCHAR(10) NOT NULL CHECK (sender_type IN ('user', 'admin')),
    sender_id UUID NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_chat_order_id ON order_chat_messages(order_id);
CREATE INDEX IF NOT EXISTS idx_order_chat_created_at ON order_chat_messages(created_at);
