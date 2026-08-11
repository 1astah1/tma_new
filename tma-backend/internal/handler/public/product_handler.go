package public

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type ProductHandler struct {
	svc         *service.ProductService
	productRepo *repository.ProductRepo
}

func NewProductHandler(svc *service.ProductService, productRepo *repository.ProductRepo) *ProductHandler {
	return &ProductHandler{svc: svc, productRepo: productRepo}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.ProductFilter{
		Page:  1,
		Limit: 20,
	}

	if v := q.Get("platform"); v != "" {
		f.Platform = &v
	}
	if v := q.Get("type"); v != "" {
		f.Type = &v
	}
	if v := q.Get("section"); v != "" {
		f.Section = &v
	}
	if v := q.Get("search"); v != "" {
		f.Search = &v
	}
	if v := q.Get("min_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			f.MinPrice = &p
		}
	}
	if v := q.Get("max_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			f.MaxPrice = &p
		}
	}
	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			f.Page = p
		}
	}
	if v := q.Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			f.Limit = l
		}
	}
	if v := q.Get("sort"); v != "" {
		f.Sort = v
	}
	if v := q.Get("order"); v != "" {
		f.Order = v
	}

	// Only active products for public
	active := "active"
	f.Status = &active

	products, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": products,
		"meta": map[string]interface{}{
			"page":  f.Page,
			"limit": f.Limit,
			"total": total,
		},
	})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid product ID")
		return
	}

	product, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Product not found")
		return
	}

	type PlatformVariant struct {
		ID              uuid.UUID       `json:"id"`
		Platform        domain.Platform `json:"platform"`
		Price           float64         `json:"price"`
		DiscountPercent float64         `json:"discount_percent"`
		ImageURL        *string         `json:"image_url"`
	}

	type ProductDetail struct {
		domain.Product
		OrderCount       int               `json:"order_count"`
		PlatformVariants []PlatformVariant `json:"platform_variants,omitempty"`
	}

	detail := ProductDetail{Product: *product}
	detail.OrderCount, _ = h.productRepo.CountOrders(r.Context(), id)

	titleKey := strings.TrimSpace(product.TitleKey)
	if titleKey == "" {
		titleKey = service.NormalizeGameTitle(product.Title)
	}
	if titleKey != "" {
		variants, err := h.productRepo.ListActiveByTitleKey(r.Context(), titleKey)
		if err == nil && len(variants) > 0 {
			familyProducts := service.SelectBestFamilyProducts(variants)
			for _, variant := range familyProducts {
				detail.PlatformVariants = append(detail.PlatformVariants, PlatformVariant{
					ID:              variant.ID,
					Platform:        variant.Platform,
					Price:           variant.Price,
					DiscountPercent: variant.DiscountPercent,
					ImageURL:        variant.ImageURL,
				})
			}
			if merged := service.MergeProductPrices(familyProducts); merged != nil {
				detail.Prices = merged
			}
			if len(detail.PlatformVariants) > 1 {
				for _, variant := range variants {
					if variant.ImageURL != nil && *variant.ImageURL != "" {
						detail.ImageURL = variant.ImageURL
						break
					}
				}
			}
		}
	}

	handler.RespondJSON(w, http.StatusOK, detail)
}

func (h *ProductHandler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	platforms := []map[string]string{
		{"id": "ps4", "name": "PlayStation 4"},
		{"id": "ps5", "name": "PlayStation 5"},
		{"id": "xbox", "name": "Xbox"},
		{"id": "pc", "name": "PC"},
	}
	handler.RespondJSON(w, http.StatusOK, platforms)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		Platform        string   `json:"platform"`
		Type            string   `json:"type"`
		Price           float64  `json:"price"`
		ImageURL        string   `json:"image_url"`
		DeliveryMethods []string `json:"delivery_methods"`
	}

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	// Forward to admin handler logic - for now just validate
	if p.Title == "" || p.Price <= 0 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Title and price are required")
		return
	}

	handler.RespondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}
