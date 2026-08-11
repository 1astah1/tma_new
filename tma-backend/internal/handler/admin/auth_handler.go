package admin

import (
	"encoding/json"
	"net/http"

	"tma-backend/internal/handler"
	"tma-backend/internal/middleware"
	"tma-backend/internal/service"
)

type AuthHandler struct {
	authSvc    *service.AuthService
	bruteForce *middleware.BruteForceProtector
}

func NewAuthHandler(authSvc *service.AuthService, bruteForce *middleware.BruteForceProtector) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, bruteForce: bruteForce}
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
		if h.bruteForce != nil {
			h.bruteForce.RecordFailure(r.RemoteAddr)
		}
		handler.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		return
	}

	if h.bruteForce != nil {
		h.bruteForce.RecordSuccess(r.RemoteAddr)
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
