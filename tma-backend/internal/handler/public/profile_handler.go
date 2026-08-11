package public

import (
	"encoding/json"
	"net/http"

	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
)

type ProfileHandler struct {
	userRepo     *repository.UserRepo
	orderRepo    *repository.OrderRepo
	settingsRepo *repository.SettingsRepo
	adminRepo    *repository.AdminRepo
}

func NewProfileHandler(userRepo *repository.UserRepo, orderRepo *repository.OrderRepo, settingsRepo *repository.SettingsRepo, adminRepo *repository.AdminRepo) *ProfileHandler {
	return &ProfileHandler{userRepo: userRepo, orderRepo: orderRepo, settingsRepo: settingsRepo, adminRepo: adminRepo}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserID(r.Context())
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	// Мини-апп по этому флагу показывает владельцу раздел управления заказами.
	isAdmin := false
	var roles []string
	if h.adminRepo != nil {
		if admin, err := h.adminRepo.GetByTelegramID(r.Context(), user.TelegramID); err == nil && admin != nil && admin.IsActive {
			isAdmin = true
			roles = admin.Roles
		}
	}

	handler.RespondJSON(w, http.StatusOK, struct {
		*domain.User
		IsAdmin bool     `json:"is_admin"`
		Roles   []string `json:"roles,omitempty"`
	}{User: user, IsAdmin: isAdmin, Roles: roles})
}

func (h *ProfileHandler) GetPaymentDetails(w http.ResponseWriter, r *http.Request) {
	s, err := h.settingsRepo.Get(r.Context(), "payment_details")
	if err != nil {
		// Return defaults if not in DB
		handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"sbp":  map[string]string{"phone": "89841157865", "bank": "Альфа-Банк", "receiver": "Олеся К."},
			"card": map[string]string{"number": "2200153684839138", "bank": "Альфа-Банк"},
			"crypto": map[string]string{
				"binance": "143915969",
				"bybit":   "100543830",
				"trc20":   "TCZxsXBe8S1BiSVPEpS12UzsaxQjkHmgap",
			},
		})
		return
	}
	var details map[string]interface{}
	json.Unmarshal([]byte(s["value"].(string)), &details)
	handler.RespondJSON(w, http.StatusOK, details)
}
