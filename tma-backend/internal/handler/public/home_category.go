package public

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type homeCategorySetting struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	ImageURL    string   `json:"image_url"`
	ProductIDs  []string `json:"product_ids"`
	CatalogType string   `json:"catalog_type,omitempty"`
	Kind        string   `json:"kind,omitempty"`        // feed_section | tile
	SectionKey  string   `json:"section_key,omitempty"` // preorders | new_releases | popular
	SortOrder   int      `json:"sort_order"`
}

const (
	homeFeedPreordersID = "home-feed-preorders"
	homeFeedNewID       = "home-feed-new"
	homeFeedPopularID   = "home-feed-popular"
	homePopularLegacyID = "home-section-popular"
)

type homeCategoryListItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ImageURL     string `json:"image_url"`
	ProductCount int    `json:"product_count"`
	CatalogType  string `json:"catalog_type,omitempty"`
}

type homeCategoryDetail struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	ImageURL    string           `json:"image_url"`
	CatalogType string           `json:"catalog_type,omitempty"`
	Products    []domain.Product `json:"products"`
}

func defaultHomeCategories() []homeCategorySetting {
	return []homeCategorySetting{
		{ID: "default-game", Title: "Игры", ImageURL: "/Игры.png", CatalogType: "game", SortOrder: 0},
		{ID: "default-currency", Title: "Валюта", ImageURL: "/Валюты.png", CatalogType: "currency", SortOrder: 1},
		{ID: "default-subscription", Title: "Подписки", ImageURL: "/Подписка.png", CatalogType: "subscription", SortOrder: 2},
	}
}

func (h *ContentHandler) loadHomeCategories(r *http.Request) ([]homeCategorySetting, error) {
	setting, err := h.settingsRepo.Get(r.Context(), "home_categories")
	if err != nil {
		return defaultHomeCategories(), nil
	}

	value, _ := setting["value"].(string)
	var raw []homeCategorySetting
	if err := json.Unmarshal([]byte(value), &raw); err != nil || len(raw) == 0 {
		return defaultHomeCategories(), nil
	}

	sort.Slice(raw, func(i, j int) bool {
		if raw[i].SortOrder == raw[j].SortOrder {
			return raw[i].Title < raw[j].Title
		}
		return raw[i].SortOrder < raw[j].SortOrder
	})
	return raw, nil
}

func (h *ContentHandler) HomeCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.loadHomeCategories(r)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	items := make([]homeCategoryListItem, 0, len(categories))
	for _, cat := range categories {
		if isFeedSection(cat) {
			continue
		}
		items = append(items, homeCategoryListItem{
			ID:           cat.ID,
			Title:        cat.Title,
			ImageURL:     cat.ImageURL,
			ProductCount: len(cat.ProductIDs),
			CatalogType:  cat.CatalogType,
		})
	}
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (h *ContentHandler) HomeCategoryByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	categories, err := h.loadHomeCategories(r)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	var found *homeCategorySetting
	for i := range categories {
		if categories[i].ID == id {
			found = &categories[i]
			break
		}
	}
	if found == nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}
	if isFeedSection(*found) {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	ids := make([]uuid.UUID, 0, len(found.ProductIDs))
	for _, rawID := range found.ProductIDs {
		parsed, err := uuid.Parse(rawID)
		if err == nil {
			ids = append(ids, parsed)
		}
	}

	products := []domain.Product{}
	if len(ids) > 0 {
		list, err := h.productRepo.GetByIDs(r.Context(), ids)
		if err != nil {
			handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		for _, product := range list {
			if product.Status == domain.ProductStatusActive {
				products = append(products, product)
			}
		}
	}

	service.EnrichListProductPrices(r.Context(), h.productRepo, products)

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": homeCategoryDetail{
			ID:          found.ID,
			Title:       found.Title,
			ImageURL:    found.ImageURL,
			CatalogType: found.CatalogType,
			Products:    products,
		},
	})
}

func parseProductUUIDs(raw []string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(raw))
	for _, id := range raw {
		parsed, err := uuid.Parse(id)
		if err == nil {
			ids = append(ids, parsed)
		}
	}
	return ids
}

func isFeedSection(cat homeCategorySetting) bool {
	if cat.Kind == "feed_section" || cat.Kind == "popular_section" {
		return true
	}
	switch cat.ID {
	case homeFeedPreordersID, homeFeedNewID, homeFeedPopularID, homePopularLegacyID:
		return true
	}
	return cat.SectionKey == "preorders" || cat.SectionKey == "new_releases" || cat.SectionKey == "popular"
}

func (h *ContentHandler) loadHomeCategorySettings(ctx context.Context) []homeCategorySetting {
	setting, err := h.settingsRepo.Get(ctx, "home_categories")
	if err != nil {
		return nil
	}
	value, _ := setting["value"].(string)
	var raw []homeCategorySetting
	if json.Unmarshal([]byte(value), &raw) != nil {
		return nil
	}
	return raw
}

func feedSectionKey(cat homeCategorySetting) string {
	if cat.SectionKey != "" {
		return cat.SectionKey
	}
	switch cat.ID {
	case homeFeedPreordersID:
		return "preorders"
	case homeFeedNewID:
		return "new_releases"
	case homeFeedPopularID, homePopularLegacyID:
		return "popular"
	}
	if cat.Kind == "popular_section" {
		return "popular"
	}
	return ""
}

func (h *ContentHandler) feedSectionProductIDs(ctx context.Context, sectionKey string) []uuid.UUID {
	for _, cat := range h.loadHomeCategorySettings(ctx) {
		if feedSectionKey(cat) == sectionKey {
			if ids := parseProductUUIDs(cat.ProductIDs); len(ids) > 0 {
				return ids
			}
		}
	}
	if sectionKey == "popular" {
		legacy, err := h.settingsRepo.Get(ctx, "popular_product_ids")
		if err == nil {
			value, _ := legacy["value"].(string)
			var raw []string
			if json.Unmarshal([]byte(value), &raw) == nil {
				return parseProductUUIDs(raw)
			}
		}
	}
	return nil
}

func (h *ContentHandler) productsByOrderedIDs(ctx context.Context, ids []uuid.UUID, limit int) ([]domain.Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	list, err := h.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]domain.Product, len(list))
	for _, product := range list {
		if product.Status == domain.ProductStatusActive {
			byID[product.ID] = product
		}
	}
	ordered := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		if product, ok := byID[id]; ok {
			ordered = append(ordered, product)
		}
		if limit > 0 && len(ordered) >= limit {
			break
		}
	}
	return ordered, nil
}

type feedSectionOptions struct {
	releaseNewestFirst bool
}

func (h *ContentHandler) resolveFeedSection(ctx context.Context, sectionKey string, limit int, opts feedSectionOptions, fallback func(fetchLimit int) ([]domain.Product, error)) ([]domain.Product, error) {
	if limit <= 0 {
		limit = 12
	}
	fetchLimit := homeFeedFetchLimit(limit)
	candidatePool := homeFeedCandidatePool(limit)
	result := make([]domain.Product, 0, fetchLimit)
	seenIDs := map[string]bool{}
	seenKeys := map[string]bool{}

	if ids := h.feedSectionProductIDs(ctx, sectionKey); len(ids) > 0 {
		pinned, err := h.productsByOrderedIDs(ctx, ids, 0)
		if err != nil {
			return nil, err
		}
		for _, p := range pinned {
			if seenIDs[p.ID.String()] || isFeedProductTaken(p, seenKeys) {
				continue
			}
			result = append(result, p)
			seenIDs[p.ID.String()] = true
			registerFeedProduct(p, seenKeys)
		}
	}

	if len(result) < fetchLimit {
		auto, err := fallback(candidatePool)
		if err != nil {
			return result, err
		}
		autoCandidates := make([]domain.Product, 0, len(auto))
		for _, p := range auto {
			if service.IsPopularFeedExcluded(p.TitleKey, p.Title) {
				continue
			}
			if seenIDs[p.ID.String()] || isFeedProductTaken(p, seenKeys) {
				continue
			}
			autoCandidates = append(autoCandidates, p)
		}
		prioritizeFeedProducts(autoCandidates, opts.releaseNewestFirst)
		for _, p := range autoCandidates {
			result = append(result, p)
			seenIDs[p.ID.String()] = true
			registerFeedProduct(p, seenKeys)
			if len(result) >= fetchLimit {
				break
			}
		}
	}
	return result, nil
}

func (h *ContentHandler) resolvePreorders(ctx context.Context, limit int) ([]domain.Product, error) {
	return h.resolveFeedSection(ctx, "preorders", limit, feedSectionOptions{releaseNewestFirst: false}, func(fetchLimit int) ([]domain.Product, error) {
		return h.productRepo.ListHomeFeedSectionCandidates(ctx, "preorder", nil, fetchLimit, false)
	})
}

func (h *ContentHandler) resolveNewReleases(ctx context.Context, limit int) ([]domain.Product, error) {
	minPrice := 149.0
	return h.resolveFeedSection(ctx, "new_releases", limit, feedSectionOptions{releaseNewestFirst: true}, func(fetchLimit int) ([]domain.Product, error) {
		return h.productRepo.ListHomeFeedSectionCandidates(ctx, "new", &minPrice, fetchLimit, true)
	})
}

func (h *ContentHandler) resolvePopularProducts(ctx context.Context, limit int) ([]domain.Product, error) {
	if h.vitrinaSvc != nil {
		return h.vitrinaSvc.ListHybridPopular(ctx, limit)
	}
	return h.resolveFeedSection(ctx, "popular", limit, feedSectionOptions{releaseNewestFirst: true}, func(fetchLimit int) ([]domain.Product, error) {
		active := "active"
		gameType := "game"
		minPrice := 149.0
		listed, _, err := h.productRepo.List(ctx, repository.ProductFilter{
			Type:     &gameType,
			Status:   &active,
			MinPrice: &minPrice,
			Limit:    fetchLimit,
			Sort:     "vitrina_score",
			Order:    "desc",
		})
		return listed, err
	})
}

func floatPtr(value float64) *float64 {
	return &value
}
