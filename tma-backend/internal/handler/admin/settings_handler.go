package admin

import (
	"encoding/json"
	"net/http"

	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
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
	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
