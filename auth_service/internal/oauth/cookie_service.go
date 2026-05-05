package oauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OAuthSession struct {
	State string `json:"state"`
	Nonce string `json:"nonce"`
}

type CookieService struct {
	aead          cipher.AEAD
	secureCookies bool
}

func NewCookieService(secret string, secureCookies bool) (*CookieService, error) {
	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("oidc cookie: invalid secret length (must be 32 bytes): %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("oidc cookie: failed to create GCM: %w", err)
	}

	return &CookieService{
		aead:          aead,
		secureCookies: secureCookies,
	}, nil
}

func (cs *CookieService) SetOAuthCookie(w http.ResponseWriter, session OAuthSession) error {
	plaintext, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("oidc cookie: failed to marshal session: %w", err)
	}

	nonce := make([]byte, cs.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("oidc cookie: failed to generate nonce: %w", err)
	}

	ciphertext := cs.aead.Seal(nonce, nonce, plaintext, nil)
	encoded := base64.RawURLEncoding.EncodeToString(ciphertext)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_session",
		Value:    encoded,
		Path:     "/authService/api/v1/auth", // covers both /google/login and /google/callback (and future /github/*)
		MaxAge:   600,
		HttpOnly: true,
		Secure:   cs.secureCookies, // false in dev, true in prod
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (cs *CookieService) ReadOAuthCookie(r *http.Request) (*OAuthSession, error) {
	cookie, err := r.Cookie("oauth_session")
	if err != nil {
		return nil, fmt.Errorf("oidc cookie: missing oauth_session cookie: %w", err)
	}

	ciphertext, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("oidc cookie: failed to decode cookie: %w", err)
	}

	nonceSize := cs.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("oidc cookie: ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := cs.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("oidc cookie: failed to decrypt: %w", err)
	}

	var session OAuthSession
	if err := json.Unmarshal(plaintext, &session); err != nil {
		return nil, fmt.Errorf("oidc cookie: failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (cs *CookieService) ClearOAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_session",
		Value:    "",
		Path:     "/authService/api/v1/auth/google",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
