package middleware

import (
	"context"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/domain/access_token"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/response"
	"net/http"
	"strings"
)

type userIDKey string

const UserIDKey userIDKey = "userIDKey"

type jwtValidator interface {
	ValidateAccessToken(
		rawToken string,
	) (*access_token.AccessTokenClaims, error)
}

func AuthMiddleware(jwtValidator jwtValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.WriteError(w, http.StatusUnauthorized, response.ErrMissingTokenFromHeader,
					fmt.Sprintf("access token should be included in request header"))
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwtValidator.ValidateAccessToken(token)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, response.ErrInvalidToken,
					fmt.Sprintf("invalid access token in request header"))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
