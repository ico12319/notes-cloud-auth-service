package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type DatabaseChecker interface {
	PingContext(ctx context.Context) error
}

type HealthHandler struct {
	db DatabaseChecker
}

func NewHealthHandler(db DatabaseChecker) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/healthz", h.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/readyz", h.Readyz).Methods(http.MethodGet)
}

type healthResponse struct {
	Status string `json:"status"`
}

type readyResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	allReady := true

	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allReady = false
		} else {
			checks["database"] = "ok"
		}
	}

	resp := readyResponse{
		Checks: checks,
	}

	if allReady {
		resp.Status = "ok"
		json.NewEncoder(w).Encode(resp)
	} else {
		resp.Status = "not ready"
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
	}
}
