package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"tma-backend/internal/handler"
	"tma-backend/internal/service"
)

func UserAuth(authSvc *service.AuthService) func(http.Handler) http.Handler {
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

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), handler.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminAuth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing token")
				return
			}

			claims, err := authSvc.ValidateAdminToken(tokenStr)
			if err != nil {
				handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
				return
			}

			adminID, err := uuid.Parse(claims.AdminID)
			if err != nil {
				handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), handler.AdminIDKey, adminID)
			ctx = context.WithValue(ctx, handler.RolesKey, claims.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles := handler.GetRoles(r.Context())

			hasRole := false
			for _, required := range roles {
				for _, userRole := range userRoles {
					if userRole == required {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				handler.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	return ""
}
