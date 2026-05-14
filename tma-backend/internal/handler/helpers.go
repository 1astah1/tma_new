package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"tma-backend/internal/domain"
)

type contextKey string

const (
	UserIDKey  contextKey = "user_id"
	AdminIDKey contextKey = "admin_id"
	RolesKey   contextKey = "roles"
)

func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetAdminID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(AdminIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetRoles(ctx context.Context) []string {
	if roles, ok := ctx.Value(RolesKey).([]string); ok {
		return roles
	}
	return []string{}
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondJSON(w, status, map[string]interface{}{
		"error": &domain.APIError{
			Code:    code,
			Message: message,
		},
	})
}
