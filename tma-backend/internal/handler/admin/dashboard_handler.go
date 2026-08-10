package admin

import (
	"net/http"

	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type DashboardHandler struct {
	orderSvc   *service.OrderService
	importRepo *repository.CatalogImportRepo
}

func NewDashboardHandler(orderSvc *service.OrderService, importRepo *repository.CatalogImportRepo) *DashboardHandler {
	return &DashboardHandler{orderSvc: orderSvc, importRepo: importRepo}
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.orderSvc.GetDashboardStats(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if h.importRepo != nil {
		if summary, err := h.importRepo.Summary(r.Context()); err == nil && summary != nil {
			stats["catalog"] = summary
			stats["pending_imports"] = summary.PendingImports
		}
	}
	handler.RespondJSON(w, http.StatusOK, stats)
}
