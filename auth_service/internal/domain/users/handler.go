package users

import (
	"context"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	http_helpers "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/http"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"log"
	"net/http"
	"time"
)

type structValidator interface {
	Struct(v any) error
}

type userService interface {
	Create(ctx context.Context, request *request_models.RegisterRequest) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
}

type userResponseConverter interface {
	ToUserResponse(user *models.User) *UserResponse
}

type passwordValidator interface {
	ValidatePassword(password *string) error
}

type UserResponse struct {
	ID          string     `json:"id"`
	DisplayName string     `json:"displayName"`
	Email       string     `json:"email"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type handler struct {
	transact              database.Transactioner
	structValidator       structValidator
	userService           userService
	userResponseConverter userResponseConverter
	passwordValidator     passwordValidator
}

func NewHandler(
	transact database.Transactioner,
	structValidator structValidator,
	userService userService,
	userResponseConverter userResponseConverter,
	passwordValidator passwordValidator) *handler {
	return &handler{
		transact:              transact,
		structValidator:       structValidator,
		userService:           userService,
		userResponseConverter: userResponseConverter,
		passwordValidator:     passwordValidator,
	}
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Register endpoint hit")

	var request request_models.RegisterRequest
	if err := http_helpers.DecodeRequestBody(w, r, &request); err != nil {
		return
	}

	if err := h.passwordValidator.ValidatePassword(request.Password); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrCodeInvalidLoginCredentials, err.Error())
		return
	}

	if err := h.structValidator.Struct(request); err != nil {
		log.Printf("validation failed for register request: %s", err.Error())
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrCodeValidationFailed, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.Printf("failed to begin transaction during user registration %s", err.Error())
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	registeredUser, err := h.userService.Create(ctx, &request)
	if err != nil {
		if database.IsUniqueViolation(err) {
			http_helpers.WriteErrorResponse(w, http.StatusConflict, http_helpers.ErrCodeEmailAlreadyExists,
				fmt.Sprintf("user with email %s already exists", request.Email))

			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("failed to commit transaction during user registration %s", err.Error())
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	log.Printf("user with id %s and email %s successfullly registered", registeredUser.ID, registeredUser.Email)

	http_helpers.WriteSuccessResponse(w, http.StatusCreated, h.userResponseConverter.ToUserResponse(registeredUser))
}

func (h *handler) Me(w http.ResponseWriter, r *http.Request) {
	log.Println("me endpoint hit")

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	userIDValue := ctx.Value(middleware.UserIDKey)
	userID, ok := userIDValue.(string)

	if !ok || userID == "" {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrCodeUnauthorized, "missing user ID")
		return
	}

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError,
			http_helpers.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	user, err := h.userService.FindByID(ctx, userID)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusNotFound,
			http_helpers.ErrCodeUserNotFound, "user not found")
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError,
			http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, h.userResponseConverter.ToUserResponse(user))
}
