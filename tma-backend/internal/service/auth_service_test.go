package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"tma-backend/internal/config"
	"tma-backend/internal/domain"
	"tma-backend/internal/service/mocks"
)

func newTestAuthService() (*AuthService, *mocks.MockUserStore, *mocks.MockAdminStore) {
	userRepo := &mocks.MockUserStore{}
	adminRepo := &mocks.MockAdminStore{}
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-for-testing-must-be-long-enough",
			AccessTTL: 24 * time.Hour,
		},
	}
	svc := NewAuthService(cfg, userRepo, adminRepo)
	return svc, userRepo, adminRepo
}

func TestGenerateUserToken(t *testing.T) {
	svc, _, _ := newTestAuthService()

	user := &domain.User{
		ID:         uuid.New(),
		TelegramID: 123456,
		Username:   strPtr("testuser"),
	}

	token, err := svc.GenerateUserToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateUserToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID.String(), claims.UserID)
	assert.Equal(t, user.TelegramID, claims.TelegramID)
}

func TestValidateUserToken_Valid(t *testing.T) {
	svc, _, _ := newTestAuthService()

	user := &domain.User{
		ID:         uuid.New(),
		TelegramID: 789012,
	}

	token, err := svc.GenerateUserToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateUserToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID.String(), claims.UserID)
	assert.Equal(t, user.TelegramID, claims.TelegramID)
}

func TestValidateUserToken_Expired(t *testing.T) {
	svc, _, _ := newTestAuthService()

	user := &domain.User{
		ID:         uuid.New(),
		TelegramID: 111222,
	}

	token, err := svc.GenerateUserToken(user)
	require.NoError(t, err)

	_, err = svc.ValidateUserToken(token)
	require.NoError(t, err)
}

func TestValidateUserToken_Invalid(t *testing.T) {
	svc, _, _ := newTestAuthService()

	_, err := svc.ValidateUserToken("invalid-token-string")
	assert.Error(t, err)
}

func TestValidateUserToken_WrongSecret(t *testing.T) {
	svc1, _, _ := newTestAuthService()

	user := &domain.User{
		ID:         uuid.New(),
		TelegramID: 333444,
	}

	token, err := svc1.GenerateUserToken(user)
	require.NoError(t, err)

	svc2, _, _ := newTestAuthService()
	svc2.cfg.JWT.Secret = "different-secret-key-for-testing-must-be-long!!"

	_, err = svc2.ValidateUserToken(token)
	assert.Error(t, err)
}

func TestAdminLogin_Success(t *testing.T) {
	svc, _, adminRepo := newTestAuthService()

	password := "securepassword"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	hashStr := string(hash)

	admin := &domain.Admin{
		ID:           uuid.New(),
		TelegramID:   999888,
		Username:     "admin1",
		PasswordHash: &hashStr,
		IsActive:     true,
		Roles:        []string{"admin"},
	}

	ctx := context.Background()
	adminRepo.On("GetByTelegramID", ctx, int64(999888)).Return(admin, nil)

	resultAdmin, token, err := svc.AdminLogin(ctx, 999888, password)
	require.NoError(t, err)
	assert.Equal(t, admin.ID, resultAdmin.ID)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateAdminToken(token)
	require.NoError(t, err)
	assert.Equal(t, admin.ID.String(), claims.AdminID)
	assert.Equal(t, admin.Username, claims.Username)

	adminRepo.AssertExpectations(t)
}

func TestAdminLogin_WrongPassword(t *testing.T) {
	svc, _, adminRepo := newTestAuthService()

	password := "correctpassword"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	hashStr := string(hash)

	admin := &domain.Admin{
		ID:           uuid.New(),
		TelegramID:   777666,
		Username:     "admin2",
		PasswordHash: &hashStr,
		IsActive:     true,
	}

	ctx := context.Background()
	adminRepo.On("GetByTelegramID", ctx, int64(777666)).Return(admin, nil)

	_, _, err = svc.AdminLogin(ctx, 777666, "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)

	adminRepo.AssertExpectations(t)
}

func TestAdminLogin_InactiveAdmin(t *testing.T) {
	svc, _, adminRepo := newTestAuthService()

	password := "password"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	hashStr := string(hash)

	admin := &domain.Admin{
		ID:           uuid.New(),
		TelegramID:   555444,
		Username:     "inactive_admin",
		PasswordHash: &hashStr,
		IsActive:     false,
	}

	ctx := context.Background()
	adminRepo.On("GetByTelegramID", ctx, int64(555444)).Return(admin, nil)

	_, _, err = svc.AdminLogin(ctx, 555444, password)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)

	adminRepo.AssertExpectations(t)
}

func TestAdminLogin_AdminNotFound(t *testing.T) {
	svc, _, adminRepo := newTestAuthService()

	ctx := context.Background()
	adminRepo.On("GetByTelegramID", ctx, int64(123123)).Return((*domain.Admin)(nil), domain.ErrNotFound)

	_, _, err := svc.AdminLogin(ctx, 123123, "anypassword")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)

	adminRepo.AssertExpectations(t)
}

func TestAdminLogin_NoPasswordHash(t *testing.T) {
	svc, _, adminRepo := newTestAuthService()

	admin := &domain.Admin{
		ID:           uuid.New(),
		TelegramID:   222111,
		Username:     "nopass_admin",
		PasswordHash: nil,
		IsActive:     true,
	}

	ctx := context.Background()
	adminRepo.On("GetByTelegramID", ctx, int64(222111)).Return(admin, nil)

	_, token, err := svc.AdminLogin(ctx, 222111, "")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	adminRepo.AssertExpectations(t)
}

func TestAuthenticateUser(t *testing.T) {
	svc, userRepo, _ := newTestAuthService()

	user := &domain.User{
		ID:         uuid.New(),
		TelegramID: 123456789,
		Username:   strPtr("test_user"),
	}

	ctx := context.Background()
	userRepo.On("Upsert", ctx, int64(123456789), mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(user, nil)

	result, err := svc.AuthenticateUser(ctx, "test", true)
	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)

	userRepo.AssertExpectations(t)
}

func TestValidateAdminToken_Valid(t *testing.T) {
	svc, _, adminRepo := newTestAuthService()

	admin := &domain.Admin{
		ID:         uuid.New(),
		TelegramID: 888777,
		Username:   "token_admin",
		IsActive:   true,
		Roles:      []string{"admin", "moderator"},
	}

	ctx := context.Background()
	adminRepo.On("GetByTelegramID", ctx, int64(888777)).Return(admin, nil)

	_, token, err := svc.AdminLogin(ctx, 888777, "")
	require.NoError(t, err)

	claims, err := svc.ValidateAdminToken(token)
	require.NoError(t, err)
	assert.Equal(t, admin.ID.String(), claims.AdminID)
	assert.Equal(t, admin.Username, claims.Username)
	assert.Len(t, claims.Roles, 2)
	assert.Contains(t, claims.Roles, "admin")
	assert.Contains(t, claims.Roles, "moderator")

	adminRepo.AssertExpectations(t)
}

func TestValidateAdminToken_Invalid(t *testing.T) {
	svc, _, _ := newTestAuthService()

	_, err := svc.ValidateAdminToken("invalid-admin-token")
	assert.Error(t, err)
}

func TestValidateAdminToken_Expired(t *testing.T) {
	svc, _, _ := newTestAuthService()

	claims := AdminClaims{
		AdminID:  uuid.New().String(),
		Username: "expired_admin",
		Roles:    []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(svc.cfg.JWT.Secret))
	require.NoError(t, err)

	_, err = svc.ValidateAdminToken(tokenStr)
	assert.Error(t, err)
}
