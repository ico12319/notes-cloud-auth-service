package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	http_helpers "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/http"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type randomReader interface {
	Read(b []byte) (n int, err error)
}

type stringEncoder interface {
	EncodeToString(b []byte) string
}

type cookieService interface {
	SetOAuthCookie(w http.ResponseWriter, session OAuthSession) error
	ReadOAuthCookie(r *http.Request) (*OAuthSession, error)
	ClearOAuthCookie(w http.ResponseWriter)
}

type idTokenVerificator interface {
	Verify(ctx context.Context, exchangedToken *oauth2.Token, session *OAuthSession) error
}

type handler struct {
	provider           *Provider
	cookieService      cookieService
	random             randomReader
	encoder            stringEncoder
	idTokenVerificator idTokenVerificator
}

func NewHandler(
	provider *Provider,
	cookieService cookieService,
	random randomReader,
	encoder stringEncoder,
	idTokenVerificator idTokenVerificator,
) *handler {
	return &handler{
		provider:           provider,
		cookieService:      cookieService,
		random:             random,
		encoder:            encoder,
		idTokenVerificator: idTokenVerificator,
	}
}

func (h *handler) Start(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /auth/google/start hit")

	state, err := h.generateRandomString(32)
	if err != nil {
		log.Printf("failed to generate state: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	nonce, err := h.generateRandomString(32)
	if err != nil {
		log.Printf("failed to generate nonce: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	codeVerifier, err := h.generateRandomString(32)
	if err != nil {
		log.Printf("failed to generate code_verifier: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	codeChallenge := h.computeCodeChallenge(codeVerifier)

	session := OAuthSession{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
	}

	if err := h.cookieService.SetOAuthCookie(w, session); err != nil {
		log.Printf("failed to set oauth cookie: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	authURL := h.provider.OAuth2Config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("access_type", "online"),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)

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

	token, err := h.provider.OAuth2Config.Exchange(
		ctx,
		authCode,
		oauth2.SetAuthURLParam("code_verifier", session.CodeVerifier),
	)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrFailedToLoginWithOIDCProvider,
			fmt.Sprintf("failed to login with google"))
		return
	}

	if err := h.idTokenVerificator.Verify(ctx, token, session); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest, http_helpers.ErrFailedToValidateIDToken, err.Error())
		return
	}

}

func (h *handler) generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := h.random.Read(b); err != nil {
		return "", err
	}
	return h.encoder.EncodeToString(b), nil
}

func (h *handler) computeCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
