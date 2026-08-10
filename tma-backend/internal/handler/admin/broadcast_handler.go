package admin

import (
	"encoding/json"
	"net/http"

	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type BroadcastHandler struct {
	userRepo *repository.UserRepo
	notifSvc *service.NotificationService
}

func NewBroadcastHandler(userRepo *repository.UserRepo, notifSvc *service.NotificationService) *BroadcastHandler {
	return &BroadcastHandler{userRepo: userRepo, notifSvc: notifSvc}
}

func (h *BroadcastHandler) Send(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}
	if req.Message == "" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "message is required")
		return
	}

	result, err := h.notifSvc.BroadcastToAllUsers(r.Context(), h.userRepo, req.Message, nil)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, result)
}
