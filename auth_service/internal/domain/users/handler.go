package users

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/response"
	"log"
	"net/http"
	"time"
)

type structValidator interface {
	Struct(v any) error
}

type userService interface {
	Create(ctx context.Context, request *request_models.RegisterRequest) (*models.User, error)
}

type handler struct {
	transact        database.Transactioner
	structValidator structValidator
	userService     userService
}

func NewHandler(
	transact database.Transactioner,
	structValidator structValidator,
	userService userService) *handler {
	return &handler{
		transact:        transact,
		structValidator: structValidator,
		userService:     userService,
	}
}

func (h *handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/authService/api/v1/register", h.Register).Methods(http.MethodPost)
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Register endpoint hit")

	var request request_models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.WriteError(w, http.StatusBadRequest, response.ErrCodeInvalidRequestBody, err.Error())
		return
	}

	if err := h.structValidator.Struct(request); err != nil {
		log.Printf("validation failed for register request: %s", err.Error())
		response.WriteError(w, http.StatusBadRequest, response.ErrCodeValidationFailed, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.Printf("failed to begin transaction during user registration %s", err.Error())
		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	registeredUser, err := h.userService.Create(ctx, &request)
	if err != nil {
		if database.IsUniqueViolation(err) {
			response.WriteError(w, http.StatusConflict, response.ErrCodeEmailAlreadyExists,
				fmt.Sprintf("user with email %s already exists", request.Email))

			return
		}

		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("failed to commit transaction during user registration %s", err.Error())
		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	log.Printf("user with id %s and email %s successfullly registered", registeredUser.ID, registeredUser.Email)

	response.WriteSuccess(w, http.StatusCreated, registeredUser)
}
