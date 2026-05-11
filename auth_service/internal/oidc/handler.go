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

type oauthSessionBuilder interface {
	Build() (*oauth.OAuthSession, error)
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
	oauthSessionBuilder   oauthSessionBuilder
	userAuthInfoExtractor userAuthInfoExtractor
	userResolver          userResolver
	tokenBundleIssuer     tokenBundleIssuer
	frontendURL           string
}

func NewHandler(
	oidcProvider OIDCProvider,
	cookieService cookieService,
	oauthSessionBuilder oauthSessionBuilder,
	userAuthInfoExtractor userAuthInfoExtractor,
	userResolver userResolver,
	tokenBundleIssuer tokenBundleIssuer,
	transact database.Transactioner,
	frontendURL string,
) *handler {
	return &handler{
		oidcProvider:          oidcProvider,
		cookieService:         cookieService,
		oauthSessionBuilder:   oauthSessionBuilder,
		userAuthInfoExtractor: userAuthInfoExtractor,
		userResolver:          userResolver,
		tokenBundleIssuer:     tokenBundleIssuer,
		transact:              transact,
		frontendURL:           frontendURL,
	}
}

func (h *handler) Start(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /auth/google/start hit")

	oauthSession, err := h.oauthSessionBuilder.Build()
	if err != nil {
		log.Printf("failed to build oauth session: %s", err.Error())

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if err := h.cookieService.SetOAuthCookie(w, *oauthSession); err != nil {
		log.Printf("failed to set oauth cookie: %v", err)
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	authURL := h.oidcProvider.WithState(oauthSession.State).WithNonce(oauthSession.Nonce).BuildAuthCodeURL()

	log.Printf("redirecting to auth url %s", authURL)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *handler) Callback(w http.ResponseWriter, r *http.Request) {
	log.Println("/callback was hit")

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	session, err := h.cookieService.ReadOAuthCookie(r)
	if err != nil {
		log.Printf("failed to read session cookie %s", err.Error())
		redirectURL := fmt.Sprintf("%s?error=invalid_session&message=%s",
			h.frontendURL, "Invalid OAuth session. Please try again.")
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	defer h.cookieService.ClearOAuthCookie(w)

	stateInURLQuery := r.URL.Query().Get("state")
	if stateInURLQuery == "" || stateInURLQuery != session.State {
		log.Printf("invalid oauth state: expected %s, got %s", session.State, stateInURLQuery)
		http.Redirect(w, r, fmt.Sprintf("%s?error=invalid_state&message=%s",
			h.frontendURL, "Invalid OAuth state. Possible CSRF attack."), http.StatusFound)

		return
	}

	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		http.Redirect(w, r, fmt.Sprintf("%s?error=no_auth_code&message=%s",
			h.frontendURL, "No authorization code received from provider."), http.StatusFound)
		return
	}

	token, err := h.oidcProvider.ExchangeAuthCodeForAccessToken(ctx, authCode)
	if err != nil {
		log.Printf("failed to exchange auth code: %v", err)
		http.Redirect(w, r, fmt.Sprintf("%s?error=token_exchange_failed&message=%s",
			h.frontendURL, "Failed to authenticate with provider. Please try again."), http.StatusFound)
		return
	}

	userAuthInfo, err := h.userAuthInfoExtractor.Extract(ctx, token, session)
	if err != nil {
		log.Printf("failed to extract user info: %v", err)
		http.Redirect(w, r, fmt.Sprintf("%s?error=invalid_token&message=%s",
			h.frontendURL, "Failed to validate authentication. Please try again."), http.StatusFound)
		return
	}

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		http.Redirect(w, r, fmt.Sprintf("%s?error=server_error&message=%s",
			h.frontendURL, "Server error occurred. Please try again."), http.StatusFound)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = database.SaveToContext(ctx, tx)

	resolvedUser, err := h.userResolver.FindOrCreateByOAuthIdentity(ctx, userAuthInfo)
	if err != nil {
		if api_errors.IsEmailAlreadyExist(err) {
			http.Redirect(w, r, fmt.Sprintf("%s?error=email_exists&message=%s",
				h.frontendURL, "This email is already registered with a different login method"), http.StatusFound)
			return
		}

		redirectURL := fmt.Sprintf("%s?error=oauth_failed&message=%s",
			h.frontendURL, "Authentication failed. Please try again.")
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	tokenBundle, err := h.tokenBundleIssuer.GenerateBundle(ctx, resolvedUser.ID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=token_failed&message=%s",
			h.frontendURL, "Failed to generate authentication token"), http.StatusFound)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s?error=server_error&message=%s",
			h.frontendURL, "Server error occurred. Please try again."), http.StatusFound)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokenBundle.AccessToken.Token,
		MaxAge:   3600,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, h.frontendURL, http.StatusFound)
}
