package auth

import (
	"context"
	"fmt"
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
		if api_errors.IsEmailNotVerified(err) {
			http_helpers.WriteErrorResponse(w, http.StatusForbidden, http_helpers.ErrEmailNotVerified,
				fmt.Sprintf("email %s is not verified", loginRequest.Email))
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeUserLoginFailed, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    loginResponse.RefreshToken,
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http_helpers.WriteSuccessResponse(w, http.StatusOK, loginResponse.AccessToken)
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("/logout endpoint hit")

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		resetRefreshTokenCookie(w)
		w.WriteHeader(http.StatusNoContent)

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

	if err := h.refreshTokenRevoker.Revoke(ctx, cookie.Value); err != nil {
		if api_errors.IsRefreshTokenNotFound(err) {
			resetRefreshTokenCookie(w)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeUserLogoutFailed, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeTransactionCommitFailed, err.Error())
		return
	}

	resetRefreshTokenCookie(w)

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) Refresh(w http.ResponseWriter, r *http.Request) {
	log.Printf("/refresh hit")

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized, http_helpers.ErrCodeUnauthorized, "refresh token cookie not found")
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

	refreshRequest := &request_models.RefreshRequest{
		RefreshToken: cookie.Value,
	}

	tokenBundle, err := h.refreshService.Refresh(ctx, refreshRequest)
	if err != nil {
		if api_errors.IsInvalidRefreshTokenError(err) {
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})
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

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenBundle.RefreshToken,
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http_helpers.WriteSuccessResponse(w, http.StatusOK, tokenBundle.AccessToken)
}

func resetRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}
