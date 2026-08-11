package public

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type PromoHandler struct {
	promoSvc *service.PromoService
}

func NewPromoHandler(promoSvc *service.PromoService) *PromoHandler {
	return &PromoHandler{promoSvc: promoSvc}
}

func (h *PromoHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code     string  `json:"code"`
		Subtotal float64 `json:"subtotal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	promo, finalPrice, err := h.promoSvc.ValidateAndApply(req.Code, req.Subtotal)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":       true,
		"code":        promo.Code,
		"discount":    req.Subtotal - finalPrice,
		"final_price": finalPrice,
	})
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
	var promo struct {
		Code            string   `json:"code"`
		DiscountPercent *float64 `json:"discount_percent"`
		DiscountFixed   *float64 `json:"discount_fixed"`
		UsageLimit      *int     `json:"usage_limit"`
		ValidUntil      *string  `json:"valid_until"`
	}
	if err := json.NewDecoder(r.Body).Decode(&promo); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	p := &repository.PromoCode{
		ID:              uuid.New(),
		Code:            promo.Code,
		DiscountPercent: 0,
		DiscountFixed:   0,
		UsageLimit:      nil,
		ValidFrom:       time.Now(),
		ValidUntil:      nil,
		IsActive:        true,
	}

	if promo.DiscountPercent != nil {
		p.DiscountPercent = *promo.DiscountPercent
	}
	if promo.DiscountFixed != nil {
		p.DiscountFixed = *promo.DiscountFixed
	}
	if promo.UsageLimit != nil {
		p.UsageLimit = promo.UsageLimit
	}
	if promo.ValidUntil != nil {
		t, err := time.Parse(time.RFC3339, *promo.ValidUntil)
		if err == nil {
			p.ValidUntil = &t
		}
	}

	if err := h.promoSvc.Create(p); err != nil {
		http.Error(w, `{"error":"failed to create promo"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(p)
}
