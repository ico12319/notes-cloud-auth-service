package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	http_helpers "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/http"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
	"log"
	"net/http"
	"time"
)

type structValidator interface {
	Struct(v any) error
}

type userService interface {
	Create(ctx context.Context, user *models.User, password *string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	VerifyUser(ctx context.Context, verificationCode string) (*models.User, error)
}

type userResponseConverter interface {
	ToUserResponse(user *models.User) *UserResponse
}

type passwordValidator interface {
	ValidatePassword(password *string) error
}

type UserResponse struct {
	ID            string     `json:"id"`
	DisplayName   string     `json:"displayName"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"emailVerified"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty"`
}

type verificationEmailSender interface {
	SendVerificationEmail(ctx context.Context, to, code string) error
}

type emailVerificationCodeGenerator interface {
	GenerateRandomString(length int) (string, error)
}

type emailVerificationTokenService interface {
	Create(ctx context.Context, verificationCode, userID string) (*models.EmailVerificationToken, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

type handler struct {
	transact                       database.Transactioner
	structValidator                structValidator
	userService                    userService
	userResponseConverter          userResponseConverter
	passwordValidator              passwordValidator
	verificationEmailSender        verificationEmailSender
	emailVerificationCodeGenerator emailVerificationCodeGenerator
	emailVerificationTokenService  emailVerificationTokenService
}

func NewHandler(
	transact database.Transactioner,
	structValidator structValidator,
	userService userService,
	userResponseConverter userResponseConverter,
	passwordValidator passwordValidator,
	verificationEmailSender verificationEmailSender,
	emailVerificationCodeGenerator emailVerificationCodeGenerator,
	emailVerificationTokenService emailVerificationTokenService) *handler {
	return &handler{
		transact:                       transact,
		structValidator:                structValidator,
		userService:                    userService,
		userResponseConverter:          userResponseConverter,
		passwordValidator:              passwordValidator,
		verificationEmailSender:        verificationEmailSender,
		emailVerificationCodeGenerator: emailVerificationCodeGenerator,
		emailVerificationTokenService:  emailVerificationTokenService,
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

	registeredUser, err := h.userService.Create(ctx, &models.User{
		Name:  request.Name,
		Email: request.Email,
	}, request.Password)
	if err != nil {
		if api_errors.IsEmailAlreadyExist(err) {
			http_helpers.WriteErrorResponse(w, http.StatusConflict, http_helpers.ErrCodeEmailAlreadyExists,
				fmt.Sprintf("user with email %s already exists", request.Email))

			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	log.Printf("user with id %s and email %s successfullly registered but email is not yet verified", registeredUser.ID, registeredUser.Email)
	log.Printf("sending verification email to user's email %s", registeredUser.Email)

	emailVerificationCode, err := h.emailVerificationCodeGenerator.GenerateRandomString(32)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if _, err := h.emailVerificationTokenService.Create(ctx, emailVerificationCode, registeredUser.ID); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("failed to commit transaction during user registration %s", err.Error())
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	go func() {
		childContext, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 30*time.Second)
		defer cancel()

		if err := h.verificationEmailSender.SendVerificationEmail(childContext, registeredUser.Email, emailVerificationCode); err != nil {
			log.Printf("failed to send verification email to email %s: %s", registeredUser.Email, err.Error())
		}
	}()

	http_helpers.WriteSuccessResponse(w, http.StatusCreated, h.userResponseConverter.ToUserResponse(registeredUser))
}

func (h *handler) Me(w http.ResponseWriter, r *http.Request) {
	log.Println("me endpoint hit")

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	userID, err := accesstoken.UserIDFromContext(ctx)
	if err != nil {
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

func (h *handler) Verify(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var emailVerificationRequest request_models.EmailVerificationRequest
	if err := http_helpers.DecodeRequestBody(w, r, &emailVerificationRequest); err != nil {
		return
	}

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	verifiedUser, err := h.userService.VerifyUser(ctx, emailVerificationRequest.VerificationCode)
	if err != nil {
		if errors.Is(err, ErrInvalidVerificationCode) {
			http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrInvalidVerificationCode, "invalid verification code")
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, h.userResponseConverter.ToUserResponse(verifiedUser))
}

func (h *handler) Resend(w http.ResponseWriter, r *http.Request) {
	var emailResendRequest request_models.VerificationEmailResendRequest
	if err := http_helpers.DecodeRequestBody(w, r, &emailResendRequest); err != nil {
		return
	}

	if emailResendRequest.Email == "" {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrCodeInvalidRequestBody,
			fmt.Sprintf("email can't be empty"))
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 30*time.Second)
		defer cancel()

		if err := h.handleResend(ctx, emailResendRequest.Email); err != nil {
			log.Printf("failed to resend verification code to email %s: %s", emailResendRequest.Email, err.Error())
		}
	}()

	http_helpers.WriteSuccessResponse(w, http.StatusAccepted, map[string]string{
		"message": "If an account with that email exists and is unverified, a new code has been sent.",
	})

}

func (h *handler) handleResend(ctx context.Context, email string) error {
	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		return err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	user, err := h.userService.FindByEmail(ctx, email)
	if err != nil {
		if api_errors.IsUserNotFoundError(err) {
			log.Printf("user with email %s does not exist, won't try to resend verification code...", email)

			return nil
		}

		return err
	}

	if user.EmailVerified {
		log.Printf("user with email %s is already verified, won't try to resend verification code..", email)

		return nil
	}

	if err := h.emailVerificationTokenService.DeleteByUserID(ctx, user.ID); err != nil {
		return err
	}

	emailVerificationCode, err := h.emailVerificationCodeGenerator.GenerateRandomString(32)
	if err != nil {
		return err
	}

	if _, err := h.emailVerificationTokenService.Create(ctx, emailVerificationCode, user.ID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return h.verificationEmailSender.SendVerificationEmail(ctx, email, emailVerificationCode)
}
