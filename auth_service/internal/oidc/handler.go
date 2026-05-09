package oidc

import (
	"context"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	http_helpers "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/http"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/oauth"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type randomGenerator interface {
	GenerateRandomString(length int) (string, error)
}

type cookieService interface {
	SetOAuthCookie(w http.ResponseWriter, session oauth.OAuthSession) error
	ReadOAuthCookie(r *http.Request) (*oauth.OAuthSession, error)
	ClearOAuthCookie(w http.ResponseWriter)
}

type userAuthInfoExtractor interface {
	Extract(ctx context.Context, token *oauth2.Token, session *oauth.OAuthSession) (*models.UserAuthInfo, error)
}

type userResolver interface {
	FindOrCreateByOAuthIdentity(ctx context.Context, userAuthInfo *models.UserAuthInfo) (*models.User, error)
}

type tokenBundleIssuer interface {
	GenerateBundle(ctx context.Context, userID string) (*models.TokenBundle, error)
}

type handler struct {
	transact              database.Transactioner
	oidcProvider          OIDCProvider
	cookieService         cookieService
	randomGenerator       randomGenerator
	userAuthInfoExtractor userAuthInfoExtractor
	userResolver          userResolver
	tokenBundleIssuer     tokenBundleIssuer
}

func NewHandler(
	oidcProvider OIDCProvider,
	cookieService cookieService,
	generator randomGenerator,
	userAuthInfoExtractor userAuthInfoExtractor,
	userResolver userResolver,
	tokenBundleIssuer tokenBundleIssuer,
	transact database.Transactioner,
) *handler {
	return &handler{
		oidcProvider:          oidcProvider,
		cookieService:         cookieService,
		randomGenerator:       generator,
		userAuthInfoExtractor: userAuthInfoExtractor,
		userResolver:          userResolver,
		tokenBundleIssuer:     tokenBundleIssuer,
		transact:              transact,
	}
}

func (h *handler) Start(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /auth/google/start hit")

	state, err := h.randomGenerator.GenerateRandomString(32)
	if err != nil {
		log.Printf("failed to generate state: %v", err)
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	nonce, err := h.randomGenerator.GenerateRandomString(32)
	if err != nil {
		log.Printf("failed to generate nonce: %v", err)
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	session := oauth.OAuthSession{
		State: state,
		Nonce: nonce,
	}

	if err := h.cookieService.SetOAuthCookie(w, session); err != nil {
		log.Printf("failed to set oauth cookie: %v", err)
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	authURL := h.oidcProvider.WithState(state).WithNonce(nonce).BuildAuthCodeURL()

	log.Printf("redirecting to auth url %s", authURL)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *handler) Callback(w http.ResponseWriter, r *http.Request) {
	log.Println("/callback was hit")

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	session, err := h.cookieService.ReadOAuthCookie(r)
	if err != nil {
		log.Printf("faied to read session cookie %s", err.Error())
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrInvalidSessionCookie, err.Error())
		return
	}

	defer h.cookieService.ClearOAuthCookie(w)

	stateInURLQuery := r.URL.Query().Get("state")
	if stateInURLQuery == "" || stateInURLQuery != session.State {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrInvalidOauthState,
			fmt.Sprintf("invalid oauth state"))
		return
	}

	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrInvalidAuthCode,
			fmt.Sprintf("invalid authorization code"))
		return
	}

	token, err := h.oidcProvider.ExchangeAuthCodeForAccessToken(ctx, authCode)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrFailedToLoginWithOIDCProvider,
			fmt.Sprintf("failed to login with provider"))
		return
	}

	userAuthInfo, err := h.userAuthInfoExtractor.Extract(ctx, token, session)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrFailedToValidateIDToken, err.Error())
		return
	}

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	resolvedUser, err := h.userResolver.FindOrCreateByOAuthIdentity(ctx, userAuthInfo)
	if err != nil {
		if api_errors.IsEmailAlreadyExist(err) {
			http_helpers.WriteErrorResponse(w, http.StatusConflict, http_helpers.ErrCodeEmailAlreadyExists, fmt.Sprintf("email %s already exists",
				userAuthInfo.Email))
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	tokenBundle, err := h.tokenBundleIssuer.GenerateBundle(ctx, resolvedUser.ID)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, tokenBundle)
}
