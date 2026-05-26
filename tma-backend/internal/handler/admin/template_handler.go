package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
)

type TemplateHandler struct {
	repo *repository.TemplateRepo
}

func NewTemplateHandler(repo *repository.TemplateRepo) *TemplateHandler {
	return &TemplateHandler{repo: repo}
}

func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	templates, err := h.repo.ListAll(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list templates")
		return
	}

	if category != "" {
		var filtered []domain.ChatTemplate
		for _, t := range templates {
			if t.Category == category {
				filtered = append(filtered, t)
			}
		}
		templates = filtered
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": templates,
	})
}

func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid template ID")
		return
	}

	t, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Template not found")
		return
	}

	handler.RespondJSON(w, http.StatusOK, t)
}

func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req domain.ChatTemplate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if req.Title == "" || req.Message == "" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Title and message are required")
		return
	}

	if req.Category == "" {
		req.Category = "general"
	}

	if err := h.repo.Create(r.Context(), &req); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create template")
		return
	}

	handler.RespondJSON(w, http.StatusCreated, req)
}

func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid template ID")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Template not found")
		return
	}

	var req domain.ChatTemplate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	req.ID = id
	req.CreatedAt = existing.CreatedAt

	if err := h.repo.Update(r.Context(), &req); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update template")
		return
	}

	handler.RespondJSON(w, http.StatusOK, req)
}

func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid template ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete template")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *TemplateHandler) ListFAQ(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListFAQAll(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list FAQ")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

func (h *TemplateHandler) GetFAQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid FAQ ID")
		return
	}

	f, err := h.repo.GetFAQByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "FAQ not found")
		return
	}

	handler.RespondJSON(w, http.StatusOK, f)
}

func (h *TemplateHandler) CreateFAQ(w http.ResponseWriter, r *http.Request) {
	var req domain.FAQItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if req.Question == "" || req.Answer == "" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Question and answer are required")
		return
	}

	if err := h.repo.CreateFAQ(r.Context(), &req); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create FAQ")
		return
	}

	handler.RespondJSON(w, http.StatusCreated, req)
}

func (h *TemplateHandler) UpdateFAQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid FAQ ID")
		return
	}

	existing, err := h.repo.GetFAQByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "FAQ not found")
		return
	}

	var req domain.FAQItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	req.ID = id
	req.CreatedAt = existing.CreatedAt

	if err := h.repo.UpdateFAQ(r.Context(), &req); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update FAQ")
		return
	}

	handler.RespondJSON(w, http.StatusOK, req)
}

func (h *TemplateHandler) DeleteFAQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid FAQ ID")
		return
	}

	if err := h.repo.DeleteFAQ(r.Context(), id); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete FAQ")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
