package users

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"net/http"
	"time"
)

type userService interface {
	GetAll(ctx context.Context) ([]*models.User, error)
}

type handler struct {
	userService userService
	transact    database.Transactioner
}

func NewHandler(userService userService, transact database.Transactioner) *handler {
	return &handler{
		userService: userService,
		transact:    transact,
	}
}

func (h *handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/authService/api/v1/users", h.Users).Methods(http.MethodGet)
}

func (h *handler) Users(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	users, err := h.userService.GetAll(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
