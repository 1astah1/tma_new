package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"tma-backend/internal/config"
	"tma-backend/internal/domain"
	"tma-backend/internal/service/mocks"
)

func TestAuthenticateUser_DevFallbackWithInvalidHash(t *testing.T) {
	userRepo := &mocks.MockUserStore{}
	cfg := &config.Config{
		JWT:      config.JWTConfig{Secret: "test-secret-key-for-testing-must-be-long-enough", AccessTTL: time.Hour},
		Telegram: config.TelegramConfig{BotToken: "123456:ABC-DEF"},
	}
	svc := NewAuthService(cfg, userRepo, &mocks.MockAdminStore{})

	initData := `user=%7B%22id%22%3A8079513329%2C%22username%22%3A%22vktn02%22%2C%22first_name%22%3A%22Test%22%7D&auth_date=1782246698&hash=invalid`
	user := &domain.User{ID: uuid.New(), TelegramID: 8079513329, Username: strPtr("vktn02")}
	ctx := context.Background()
	userRepo.On("Upsert", ctx, int64(8079513329), mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(user, nil)

	result, err := svc.AuthenticateUser(ctx, initData, true)
	require.NoError(t, err)
	assert.Equal(t, int64(8079513329), result.TelegramID)
}
