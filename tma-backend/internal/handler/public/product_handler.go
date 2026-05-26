package public

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type ProductHandler struct {
	svc       *service.ProductService
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

	type ProductWithKeys struct {
		domain.Product
		AvailableKeys int `json:"available_keys"`
	}

	result := make([]ProductWithKeys, 0, len(products))
	for _, p := range products {
		pwk := ProductWithKeys{Product: p}
		for _, dm := range p.DeliveryMethods {
			if dm == "key" {
				cnt, _ := h.productRepo.CountAvailableKeys(r.Context(), p.ID)
				pwk.AvailableKeys = cnt
				break
			}
		}
		result = append(result, pwk)
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": result,
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

	type ProductWithKeys struct {
		domain.Product
		AvailableKeys int `json:"available_keys"`
		OrderCount    int `json:"order_count"`
	}

	pwk := ProductWithKeys{Product: *product}
	for _, dm := range product.DeliveryMethods {
		if dm == "key" {
			cnt, _ := h.productRepo.CountAvailableKeys(r.Context(), id)
			pwk.AvailableKeys = cnt
			break
		}
	}
	pwk.OrderCount, _ = h.productRepo.CountOrders(r.Context(), id)

	handler.RespondJSON(w, http.StatusOK, pwk)
}

func (h *ProductHandler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	platforms := []map[string]string{
		{"id": "ps4", "name": "PlayStation 4"},
		{"id": "ps5", "name": "PlayStation 5"},
		{"id": "xbox", "name": "Xbox"},
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
