package admin

import (
	"encoding/json"
	"net/http"

	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type SettingsHandler struct {
	repo *repository.SettingsRepo
}

func NewSettingsHandler(repo *repository.SettingsRepo) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key != "" {
		s, err := h.repo.Get(r.Context(), key)
		if err != nil {
			handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Setting not found")
			return
		}
		handler.RespondJSON(w, http.StatusOK, s)
		return
	}
	all, err := h.repo.GetAll(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, all)
}

func (h *SettingsHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}
	if err := h.repo.Upsert(r.Context(), body.Key, body.Value); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if body.Key == service.PricingSettingsKey {
		if cfg, err := service.ParsePricingFormulasValue(body.Value); err == nil {
			service.ApplyPricingFormulas(cfg)
		}
	}
	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *SettingsHandler) PricingPreview(w http.ResponseWriter, r *http.Request) {
	cfg := service.GetPricingFormulas()
	rate := service.TRYToRUBRate()
	if manual := service.ManualTRYToRUBRate(); manual > 0 {
		rate = manual
	}
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"formulas": cfg,
		"try_rub_rate": rate,
		"examples": map[string]float64{
			"turkey_500_try":  service.TurkeyNominalPrice(500),
			"ukraine_1000_uah": service.UkrainePrice(1000),
			"xbox_10_usd":     service.XboxUSAPrice(10),
		},
	})
}
