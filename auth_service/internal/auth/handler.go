package auth

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	http_helpers "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/http"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"log"
	"net/http"
	"time"
)

type loginService interface {
	Login(ctx context.Context, request *request_models.LoginRequest) (*models.TokenBundle, error)
}

type refreshTokenRevoker interface {
	Revoke(ctx context.Context, rawToken string) error
}

type refreshService interface {
	Refresh(ctx context.Context, request *request_models.RefreshRequest) (*models.TokenBundle, error)
}

type handler struct {
	transact            database.Transactioner
	loginService        loginService
	refreshTokenRevoker refreshTokenRevoker
	refreshService      refreshService
}

func NewHandler(
	loginService loginService,
	transact database.Transactioner,
	refreshTokenRevoker refreshTokenRevoker,
	refreshService refreshService) *handler {
	return &handler{
		transact:            transact,
		loginService:        loginService,
		refreshTokenRevoker: refreshTokenRevoker,
		refreshService:      refreshService,
	}
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("/login endpoint hit")

	var loginRequest request_models.LoginRequest
	if err := http_helpers.DecodeRequestBody(w, r, &loginRequest); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	loginResponse, err := h.loginService.Login(ctx, &loginRequest)
	if err != nil {
		if api_errors.IsWrongLoginCredentialsError(err) {
			http_helpers.WriteErrorResponse(w, http.StatusUnauthorized, http_helpers.ErrCodeInvalidLoginCredentials, err.Error())
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeUserLoginFailed, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, loginResponse)
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("/logout endpoint hit")

	var request request_models.LogoutRequest
	if err := http_helpers.DecodeRequestBody(w, r, &request); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	if err := h.refreshTokenRevoker.Revoke(ctx, request.RefreshToken); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeUserLogoutFailed, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) Refresh(w http.ResponseWriter, r *http.Request) {
	log.Printf("/refresh hit")

	var request request_models.RefreshRequest
	if err := http_helpers.DecodeRequestBody(w, r, &request); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	tokenBundle, err := h.refreshService.Refresh(ctx, &request)
	if err != nil {
		if api_errors.IsInvalidRefreshTokenError(err) {
			http_helpers.WriteErrorResponse(w, http.StatusUnauthorized, http_helpers.ErrCodeUnauthorized, err.Error())
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, tokenBundle)
}
