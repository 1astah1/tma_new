package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type PromoHandler struct {
	promoSvc *service.PromoService
}

func NewPromoHandler(promoSvc *service.PromoService) *PromoHandler {
	return &PromoHandler{promoSvc: promoSvc}
}

func (h *PromoHandler) List(w http.ResponseWriter, r *http.Request) {
	promos, err := h.promoSvc.List()
	if err != nil {
		http.Error(w, `{"error":"failed to list promos"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": promos})
}

func (h *PromoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code            string   `json:"code"`
		DiscountPercent *float64 `json:"discount_percent"`
		DiscountFixed   *float64 `json:"discount_fixed"`
		UsageLimit      *int     `json:"usage_limit"`
		ValidUntil      *string  `json:"valid_until"`
		IsActive        *bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	p := &repository.PromoCode{
		ID:              uuid.New(),
		Code:            req.Code,
		DiscountPercent: 0,
		DiscountFixed:   0,
		UsageLimit:      nil,
		ValidFrom:       time.Now(),
		ValidUntil:      nil,
		IsActive:        true,
	}

	if req.DiscountPercent != nil {
		p.DiscountPercent = *req.DiscountPercent
	}
	if req.DiscountFixed != nil {
		p.DiscountFixed = *req.DiscountFixed
	}
	if req.UsageLimit != nil {
		p.UsageLimit = req.UsageLimit
	}
	if req.ValidUntil != nil {
		t, err := time.Parse(time.RFC3339, *req.ValidUntil)
		if err == nil {
			p.ValidUntil = &t
		}
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}

	if err := h.promoSvc.Create(p); err != nil {
		http.Error(w, `{"error":"failed to create promo"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(p)
}

func (h *PromoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid promo ID")
		return
	}

	promo, err := h.promoSvc.GetByID(id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Promo not found")
		return
	}

	handler.RespondJSON(w, http.StatusOK, promo)
}

func (h *PromoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid promo ID")
		return
	}

	var req struct {
		Code            *string  `json:"code"`
		DiscountPercent *float64 `json:"discount_percent"`
		DiscountFixed   *float64 `json:"discount_fixed"`
		UsageLimit      *int     `json:"usage_limit"`
		ValidUntil      *string  `json:"valid_until"`
		IsActive        *bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	promo, err := h.promoSvc.GetByID(id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Promo not found")
		return
	}

	if req.Code != nil {
		promo.Code = *req.Code
	}
	if req.DiscountPercent != nil {
		promo.DiscountPercent = *req.DiscountPercent
	}
	if req.DiscountFixed != nil {
		promo.DiscountFixed = *req.DiscountFixed
	}
	if req.UsageLimit != nil {
		promo.UsageLimit = req.UsageLimit
	}
	if req.ValidUntil != nil {
		t, err := time.Parse(time.RFC3339, *req.ValidUntil)
		if err == nil {
			promo.ValidUntil = &t
		}
	}
	if req.IsActive != nil {
		promo.IsActive = *req.IsActive
	}

	if err := h.promoSvc.UpdatePromo(promo); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update promo")
		return
	}

	handler.RespondJSON(w, http.StatusOK, promo)
}

func (h *PromoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid promo ID")
		return
	}

	if err := h.promoSvc.Delete(id.String()); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete promo")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
