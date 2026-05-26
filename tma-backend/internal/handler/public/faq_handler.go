package public

import (
	"net/http"

	"github.com/google/uuid"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
)

type FAQHandler struct {
	repo *repository.TemplateRepo
}

func NewFAQHandler(repo *repository.TemplateRepo) *FAQHandler {
	return &FAQHandler{repo: repo}
}

func (h *FAQHandler) ListFAQ(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListFAQ(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list FAQ")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

func (h *FAQHandler) GetFAQAnswer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid FAQ ID")
		return
	}

	f, err := h.repo.GetFAQByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "FAQ not found")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"question": f.Question,
		"answer":   f.Answer,
	})
}
