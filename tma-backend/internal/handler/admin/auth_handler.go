package admin

import (
	"encoding/json"
	"net/http"

	"tma-backend/internal/handler"
	"tma-backend/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	admin, token, err := h.authSvc.AdminLogin(r.Context(), req.TelegramID, req.Password)
	if err != nil {
		handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"admin": admin,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	adminID := handler.GetAdminID(r.Context())
	roles := handler.GetRoles(r.Context())

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"admin_id": adminID,
		"roles":    roles,
	})
}
