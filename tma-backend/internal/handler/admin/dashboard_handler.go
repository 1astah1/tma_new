package admin

import (
	"net/http"

	"tma-backend/internal/handler"
	"tma-backend/internal/service"
)

type DashboardHandler struct {
	orderSvc *service.OrderService
}

func NewDashboardHandler(orderSvc *service.OrderService) *DashboardHandler {
	return &DashboardHandler{orderSvc: orderSvc}
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.orderSvc.GetDashboardStats(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, stats)
}
