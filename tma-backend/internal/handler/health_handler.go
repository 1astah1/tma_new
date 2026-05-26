package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct {
	db DBChecker
}

type DBChecker interface {
	PingContext(ctx context.Context) error
}

func NewHealthHandler(db DBChecker) *HealthHandler {
	return &HealthHandler{db: db}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": "database connection failed"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
