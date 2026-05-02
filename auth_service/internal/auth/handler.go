package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/response"
	"log"
	"net/http"
	"time"
)

type loginService interface {
	Login(ctx context.Context, request *request_models.LoginRequest) (*models.LoginResponse, error)
}

type refreshTokenRevoker interface {
	Revoke(ctx context.Context, rawToken string) error
}

type handler struct {
	transact            database.Transactioner
	loginService        loginService
	refreshTokenRevoker refreshTokenRevoker
}

func NewHandler(
	loginService loginService,
	transact database.Transactioner,
	refreshTokenRevoker refreshTokenRevoker) *handler {
	return &handler{
		transact:            transact,
		loginService:        loginService,
		refreshTokenRevoker: refreshTokenRevoker,
	}
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("login endpoint hit")

	var loginRequest request_models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		response.WriteError(w, http.StatusBadRequest, response.ErrCodeInvalidRequestBody, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	loginResponse, err := h.loginService.Login(ctx, &loginRequest)
	if err != nil {
		if errors.Is(err, ErrWrongLoginCredentials) {
			response.WriteError(w, http.StatusUnauthorized, response.ErrCodeInvalidLoginCredentials, ErrWrongLoginCredentials.Error())
			return
		}

		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeUserLoginFailed, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, loginResponse)
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("logout endpoint hit")

	var request request_models.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.WriteError(w, http.StatusBadRequest, response.ErrCodeInvalidRequestBody, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeTransactionBeginFailed, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	if err := h.refreshTokenRevoker.Revoke(ctx, request.RefreshToken); err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeUserLogoutFailed, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
