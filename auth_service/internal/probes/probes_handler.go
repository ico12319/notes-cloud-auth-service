package probes

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type healthResponse struct {
	Status string `json:"status"`
}

func (h *healthResponse) SetStatus(status string) {
	h.Status = status
}

type readyResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func (r *readyResponse) SetStatus(status string) {
	r.Status = status
}

func (r *readyResponse) SetChecks(checks map[string]string) {
	r.Checks = checks
}

type databaseChecker interface {
	PingContext(ctx context.Context) error
}

type healthHandler struct {
	databaseChecker databaseChecker
}

func NewHealthHandler(databaseChecker databaseChecker) *healthHandler {
	return &healthHandler{databaseChecker: databaseChecker}
}

func (h *healthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/authService/api/v1/healthz", h.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/authService/api/v1/readyz", h.Readyz).Methods(http.MethodGet)
}

func (h *healthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(healthResponse{Status: "ok"}); err != nil {
		http.Error(w, "failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

func (h *healthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	allReady := true

	if h.databaseChecker != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := h.databaseChecker.PingContext(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allReady = false
		} else {
			checks["database"] = "healthy"
		}
	}

	resp := &readyResponse{}
	resp.SetChecks(checks)

	if allReady {
		resp.SetStatus("ok")
	} else {
		resp.SetStatus("not ready")
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}
