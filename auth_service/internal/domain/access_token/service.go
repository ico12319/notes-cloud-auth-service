package access_token

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"time"
)

const tokenTypeBearer = "Bearer"

type timeService interface {
	Now() time.Time
}

type AccessTokenClaims struct {
	UserID string `json:"userId"`

	jwt.RegisteredClaims
}

type service struct {
	timeService        timeService
	cfg                config.AccessToken
	tokenSigningMethod jwt.SigningMethod
}

func NewService(
	timeService timeService,
	cfg config.AccessToken,
	tokenSigningMethod jwt.SigningMethod) *service {
	return &service{
		timeService:        timeService,
		cfg:                cfg,
		tokenSigningMethod: tokenSigningMethod,
	}
}

func (s *service) GenerateForUser(
	userID string,
) (*models.AccessToken, error) {
	now := s.timeService.Now().UTC()
	expiresAt := now.Add(s.cfg.TTL)

	claims := AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.cfg.Issuer,
			Audience:  jwt.ClaimStrings{s.cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signedToken, err := jwt.NewWithClaims(s.tokenSigningMethod, claims).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	return &models.AccessToken{
		Token:     signedToken,
		TokenType: tokenTypeBearer,
		ExpiresIn: int(s.cfg.TTL.Seconds()),
		ExpiresAt: expiresAt,
	}, nil
}

func (s *service) ValidateAccessToken(
	rawToken string,
) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		rawToken,
		&AccessTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != s.tokenSigningMethod {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}

			return []byte(s.cfg.Secret), nil
		},
		jwt.WithIssuer(s.cfg.Issuer),
		jwt.WithAudience(s.cfg.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid access token claims")
	}

	return claims, nil
}
