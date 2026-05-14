package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"tma-backend/internal/config"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type AuthService struct {
	cfg    *config.Config
	userRepo *repository.UserRepo
	adminRepo *repository.AdminRepo
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepo, adminRepo *repository.AdminRepo) *AuthService {
	return &AuthService{cfg: cfg, userRepo: userRepo, adminRepo: adminRepo}
}

type UserClaims struct {
	UserID    string `json:"user_id"`
	TelegramID int64 `json:"telegram_id"`
	jwt.RegisteredClaims
}

type AdminClaims struct {
	AdminID  string   `json:"admin_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func (s *AuthService) VerifyTelegramInitData(initData string) (bool, error) {
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return false, err
	}

	hash := parsed.Get("hash")
	parsed.Del("hash")

	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckStrings []string
	for _, k := range keys {
		dataCheckStrings = append(dataCheckStrings, fmt.Sprintf("%s=%s", k, parsed.Get(k)))
	}
	dataCheck := strings.Join(dataCheckStrings, "\n")

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(s.cfg.Telegram.BotToken))
	secretKey := secret.Sum(nil)

	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheck))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(hash), []byte(expectedHash)), nil
}

func (s *AuthService) AuthenticateUser(ctx context.Context, initData string) (*domain.User, error) {
	tgID, username, firstName := extractTelegramData(initData)

	user, err := s.userRepo.Upsert(ctx, tgID, username, firstName)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) GenerateUserToken(user *domain.User) (string, error) {
	claims := UserClaims{
		UserID:     user.ID.String(),
		TelegramID: user.TelegramID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWT.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *AuthService) ValidateUserToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

func (s *AuthService) AdminLogin(ctx context.Context, telegramID int64, password string) (*domain.Admin, string, error) {
	admin, err := s.adminRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, "", domain.ErrUnauthorized
	}
	if !admin.IsActive {
		return nil, "", domain.ErrForbidden
	}

	claims := AdminClaims{
		AdminID:  admin.ID.String(),
		Username: admin.Username,
		Roles:    admin.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWT.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, "", err
	}

	return admin, tokenStr, nil
}

func (s *AuthService) ValidateAdminToken(tokenStr string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

func extractTelegramData(initData string) (telegramID int64, username, firstName *string) {
	parsed, _ := url.ParseQuery(initData)

	if id := parsed.Get("id"); id != "" {
		fmt.Sscanf(id, "%d", &telegramID)
	}

	u := parsed.Get("username")
	if u != "" {
		username = &u
	}
	fn := parsed.Get("first_name")
	if fn != "" {
		firstName = &fn
	}
	return
}
