package public

import (
	"context"
	"net/http"

	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type ContentHandler struct {
	productRepo  *repository.ProductRepo
	settingsRepo *repository.SettingsRepo
	vitrinaSvc   VitrinaPopularLister
}

type VitrinaPopularLister interface {
	ListHybridPopular(ctx context.Context, limit int) ([]domain.Product, error)
}

func NewContentHandler(productRepo *repository.ProductRepo, settingsRepo *repository.SettingsRepo, vitrinaSvc VitrinaPopularLister) *ContentHandler {
	return &ContentHandler{productRepo: productRepo, settingsRepo: settingsRepo, vitrinaSvc: vitrinaSvc}
}

func (h *ContentHandler) PopularProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.resolvePopularProducts(r.Context(), 20)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	service.EnrichListProductPrices(r.Context(), h.productRepo, products)
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{"data": products})
}

func (h *ContentHandler) HomeFeed(w http.ResponseWriter, r *http.Request) {
	const limit = 12
	fetchLimit := homeFeedFetchLimit(limit)

	preordersRaw, err := h.resolvePreorders(r.Context(), fetchLimit)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	newRaw, err := h.resolveNewReleases(r.Context(), fetchLimit)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	popularRaw, err := h.resolvePopularProducts(r.Context(), fetchLimit)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	preorders, newReleases, popular := assembleHomeFeedSections(preordersRaw, newRaw, popularRaw, limit)

	service.EnrichListProductPrices(r.Context(), h.productRepo, preorders)
	service.EnrichListProductPrices(r.Context(), h.productRepo, newReleases)
	service.EnrichListProductPrices(r.Context(), h.productRepo, popular)

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"preorders":    preorders,
			"new_releases": newReleases,
			"popular":      popular,
		},
	})
}

func (h *ContentHandler) ShopSettings(w http.ResponseWriter, r *http.Request) {
	result := map[string]interface{}{
		"support_url":  "https://t.me/coin_mint_chat",
		"reviews_url":  "https://t.me/coin_mint_reviews",
		"shop_rules":   "",
	}

	for _, key := range []string{"support_url", "reviews_url", "shop_rules"} {
		setting, err := h.settingsRepo.Get(r.Context(), key)
		if err != nil {
			continue
		}
		if value, ok := setting["value"].(string); ok && value != "" {
			result[key] = value
		}
	}

	handler.RespondJSON(w, http.StatusOK, result)
}

func stringPtr(value string) *string {
	return &value
}
