package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"tma-backend/internal/config"
	"tma-backend/internal/domain"
)

type AuthService struct {
	cfg       *config.Config
	userRepo  UserStore
	adminRepo AdminStore
}

func NewAuthService(cfg *config.Config, userRepo UserStore, adminRepo AdminStore) *AuthService {
	return &AuthService{cfg: cfg, userRepo: userRepo, adminRepo: adminRepo}
}

type UserClaims struct {
	UserID     string `json:"user_id"`
	TelegramID int64  `json:"telegram_id"`
	jwt.RegisteredClaims
}

type AdminClaims struct {
	AdminID  string   `json:"admin_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func (s *AuthService) VerifyTelegramInitData(initData string) (bool, error) {
	if s.cfg.Telegram.BotToken == "" {
		return false, domain.ErrUnauthorized
	}

	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return false, err
	}

	hash := parsed.Get("hash")
	if hash == "" {
		return false, domain.ErrUnauthorized
	}
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

	if !hmac.Equal([]byte(hash), []byte(expectedHash)) {
		return false, domain.ErrUnauthorized
	}

	if authDate := parsed.Get("auth_date"); authDate != "" {
		ts, err := strconv.ParseInt(authDate, 10, 64)
		if err == nil && time.Since(time.Unix(ts, 0)) > 24*time.Hour {
			return false, domain.ErrUnauthorized
		}
	}

	return true, nil
}

func (s *AuthService) AuthenticateUser(ctx context.Context, initData string, devBypass bool) (*domain.User, error) {
	if devBypass && (initData == "" || initData == "test") {
		tgID := int64(123456789)
		username := "test_user"
		firstName := "Test"
		user, err := s.userRepo.Upsert(ctx, tgID, &username, &firstName)
		if err != nil {
			return nil, err
		}
		if user.IsBanned {
			return nil, domain.ErrForbidden
		}
		return user, nil
	}

	if initData == "" {
		return nil, domain.ErrUnauthorized
	}

	valid, err := s.VerifyTelegramInitData(initData)
	if err != nil || !valid {
		if devBypass {
			tgID, username, firstName := extractTelegramData(initData)
			if tgID != 0 {
				slog.Warn("dev auth: accepting Telegram initData without hash verification",
					slog.Int64("telegram_id", tgID),
				)
				user, upsertErr := s.userRepo.Upsert(ctx, tgID, username, firstName)
				if upsertErr != nil {
					return nil, upsertErr
				}
				if user.IsBanned {
					return nil, domain.ErrForbidden
				}
				return user, nil
			}
		}
		return nil, domain.ErrUnauthorized
	}

	tgID, username, firstName := extractTelegramData(initData)
	if tgID == 0 {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.userRepo.Upsert(ctx, tgID, username, firstName)
	if err != nil {
		return nil, err
	}
	if user.IsBanned {
		return nil, domain.ErrForbidden
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

	if admin.PasswordHash != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*admin.PasswordHash), []byte(password)); err != nil {
			return nil, "", domain.ErrUnauthorized
		}
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

	if userJSON := parsed.Get("user"); userJSON != "" {
		var u struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
		}
		if json.Unmarshal([]byte(userJSON), &u) == nil && u.ID != 0 {
			telegramID = u.ID
			if u.Username != "" {
				username = &u.Username
			}
			if u.FirstName != "" {
				firstName = &u.FirstName
			}
			return
		}
	}

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
