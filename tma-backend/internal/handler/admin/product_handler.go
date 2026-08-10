package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type AdminProductHandler struct {
	svc         *service.ProductService
	productRepo *repository.ProductRepo
	importRepo  *repository.CatalogImportRepo
}

func NewAdminProductHandler(svc *service.ProductService, productRepo *repository.ProductRepo, importRepo *repository.CatalogImportRepo) *AdminProductHandler {
	return &AdminProductHandler{svc: svc, productRepo: productRepo, importRepo: importRepo}
}

func (h *AdminProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.ProductFilter{Page: 1, Limit: 20}

	if v := q.Get("platform"); v != "" {
		f.Platform = &v
	}
	if v := q.Get("type"); v != "" {
		f.Type = &v
	}
	if v := q.Get("search"); v != "" {
		f.Search = &v
	}
	if v := q.Get("status"); v != "" {
		f.Status = &v
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

	products, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	type ProductWithCounts struct {
		domain.Product
		OrderCount int `json:"order_count"`
	}

	result := make([]ProductWithCounts, 0, len(products))
	for _, p := range products {
		pwc := ProductWithCounts{Product: p}
		pwc.OrderCount, _ = h.productRepo.CountOrders(r.Context(), p.ID)
		result = append(result, pwc)
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

func (h *AdminProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid ID")
		return
	}
	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Product not found")
		return
	}
	resp := map[string]interface{}{}
	b, _ := json.Marshal(p)
	_ = json.Unmarshal(b, &resp)
	if h.importRepo != nil && p.Type == domain.ProductTypeGame {
		if imp, err := h.importRepo.GetByProductID(r.Context(), id); err == nil && imp != nil {
			resp["catalog_import"] = imp
		}
	}
	handler.RespondJSON(w, http.StatusOK, resp)
}

func (h *AdminProductHandler) SyncFromImport(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid ID")
		return
	}
	if h.importRepo == nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Import repo unavailable")
		return
	}
	imp, err := h.importRepo.GetByProductID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Связанный импорт не найден")
		return
	}
	product, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Product not found")
		return
	}
	price := service.EffectiveCatalogPriceRUB(imp)
	if imp.OriginalPriceRUB != nil && *imp.OriginalPriceRUB >= service.MinPaidPriceRUB() {
		price = *imp.OriginalPriceRUB
	}
	if price < service.MinPaidPriceRUB() && product.GameSection != "preorder" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_PRICE", "Цена в импорте ниже минимума")
		return
	}
	product.Price = price
	if len(imp.Prices) > 0 {
		product.Prices = imp.Prices
	}
	if imp.ReleaseDate != nil {
		product.ReleaseDate = imp.ReleaseDate
	}
	if imp.GameSection != "" {
		product.GameSection = imp.GameSection
	}
	if imp.ImageURL != nil && *imp.ImageURL != "" {
		product.ImageURL = imp.ImageURL
	}
	if err := h.svc.Update(r.Context(), product); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "synced",
		"price":   product.Price,
		"product": product,
	})
}

func (h *AdminProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p domain.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}
	if err := h.svc.Create(r.Context(), &p); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusCreated, p)
}

func (h *AdminProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid ID")
		return
	}
	existing, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Product not found")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(existing); err != nil {
		slog.Warn("Product update decode error", slog.String("error", err.Error()))
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}
	existing.ID = id

	if err := h.svc.Update(r.Context(), existing); err != nil {
		slog.Error("Product update error", slog.String("error", err.Error()))
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, existing)
}

func (h *AdminProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid ID")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
