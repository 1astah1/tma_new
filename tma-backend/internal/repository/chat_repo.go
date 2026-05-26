package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type ChatRepo struct {
	db *sqlx.DB
}

func NewChatRepo(db *sqlx.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) Create(ctx context.Context, msg *domain.ChatMessage) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO order_chat_messages (order_id, sender_type, sender_id, message)
		 VALUES (:order_id, :sender_type, :sender_id, :message)`, msg)
	return err
}

func (r *ChatRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage
	err := r.db.SelectContext(ctx, &messages,
		"SELECT * FROM order_chat_messages WHERE order_id=$1 ORDER BY created_at ASC", orderID)
	if err != nil {
		return nil, err
	}
	if messages == nil {
		messages = []domain.ChatMessage{}
	}
	return messages, nil
}
