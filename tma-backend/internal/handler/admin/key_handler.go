package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type KeyHandler struct {
	keyRepo *repository.KeyRepo
}

func NewKeyHandler(keyRepo *repository.KeyRepo) *KeyHandler {
	return &KeyHandler{keyRepo: keyRepo}
}

func (h *KeyHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	keys, total, err := h.keyRepo.ListAll(ctx, &status, page, limit)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list keys")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": keys,
		"meta": map[string]int{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *KeyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid key ID")
		return
	}

	key, err := h.keyRepo.GetByID(ctx, id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Key not found")
		return
	}

	handler.RespondJSON(w, http.StatusOK, key)
}

func (h *KeyHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid key ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if req.Status != "available" && req.Status != "sold" && req.Status != "reserved" && req.Status != "invalid" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_STATUS", "Invalid status")
		return
	}

	if err := h.keyRepo.UpdateStatus(ctx, id, req.Status); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update status")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *KeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid key ID")
		return
	}

	if err := h.keyRepo.Delete(ctx, id); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete key")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *KeyHandler) BulkImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		ProductID string   `json:"product_id"`
		Keys      []string `json:"keys"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID")
		return
	}

	imported, err := h.keyRepo.BulkImport(ctx, productID, req.Keys)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to import keys")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"imported": imported,
		"total":    len(req.Keys),
	})
}

func (h *KeyHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	var ids []uuid.UUID
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	deleted, err := h.keyRepo.BulkDelete(ctx, ids)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete keys")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}

func (h *KeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		ProductID string `json:"product_id"`
		Key       string `json:"key"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if req.ProductID == "" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Product ID is required")
		return
	}

	if req.Key == "" || len(req.Key) < 5 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Key must be at least 5 characters")
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID")
		return
	}

	status := req.Status
	if status == "" {
		status = string(domain.KeyStatusAvailable)
	}

	key, err := h.keyRepo.CreateKey(ctx, productID, req.Key, domain.KeyStatus(status))
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create key")
		return
	}

	handler.RespondJSON(w, http.StatusCreated, key)
}

func (h *KeyHandler) GetByProductID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productID, err := uuid.Parse(chi.URLParam(r, "product_id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID")
		return
	}

	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	keys, total, err := h.keyRepo.GetByProductID(ctx, productID, statusPtr, page, limit)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list keys")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": keys,
		"meta": map[string]int{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *KeyHandler) GetProductKeyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productID, err := uuid.Parse(chi.URLParam(r, "product_id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID")
		return
	}

	stats, err := h.keyRepo.GetStatsByProductID(ctx, productID)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get stats")
		return
	}

	handler.RespondJSON(w, http.StatusOK, stats)
}

func (h *KeyHandler) ReleaseKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid key ID")
		return
	}

	if err := h.keyRepo.ReleaseKey(ctx, id); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to release key")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "released"})
}

func (h *KeyHandler) UpdateKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid key ID")
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if req.Key == "" || len(req.Key) < 5 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Key must be at least 5 characters")
		return
	}

	if err := h.keyRepo.UpdateKey(ctx, id, req.Key); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update key")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
