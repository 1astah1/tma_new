package admin

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

type AdminProductHandler struct {
	svc *service.ProductService
}

func NewAdminProductHandler(svc *service.ProductService) *AdminProductHandler {
	return &AdminProductHandler{svc: svc}
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

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": products,
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
	handler.RespondJSON(w, http.StatusOK, p)
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
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}
	existing.ID = id

	if err := h.svc.Update(r.Context(), existing); err != nil {
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
