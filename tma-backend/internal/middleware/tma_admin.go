package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

// TMAAdminAuth пускает в админские ручки по обычному токену мини-аппа, если
// телеграм-аккаунт есть в списке админов. Нужно, чтобы вести заказы с телефона
// прямо из Telegram, не заходя в веб-панель с паролем.
//
// Контекст заполняется теми же ключами, что и обычная админская авторизация,
// поэтому существующие обработчики работают без единой правки.
func TMAAdminAuth(
	authSvc *service.AuthService,
	userRepo *repository.UserRepo,
	adminRepo *repository.AdminRepo,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing token")
				return
			}

			claims, err := authSvc.ValidateUserToken(tokenStr)
			if err != nil {
				handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
				return
			}

			admin, err := adminRepo.GetByTelegramID(r.Context(), claims.TelegramID)
			if err != nil || admin == nil || !admin.IsActive {
				handler.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Нет доступа к админке")
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), handler.UserIDKey, userID)
			ctx = context.WithValue(ctx, handler.AdminIDKey, admin.ID)
			ctx = context.WithValue(ctx, handler.RolesKey, []string(admin.Roles))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
