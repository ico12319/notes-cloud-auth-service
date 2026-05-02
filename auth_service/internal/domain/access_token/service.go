package access_token

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"time"
)

const tokenTypeBearer = "Bearer"

type AccessTokenClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`

	jwt.RegisteredClaims
}

type timeService interface {
	Now() time.Time
}

type jwtGenerator interface {
	GenerateSignedJWT(claims jwt.Claims, secret []byte) (string, error)
}

type service struct {
	timeService  timeService
	cfg          config.AccessToken
	jwtGenerator jwtGenerator
}

func NewService(
	jwtGenerator jwtGenerator,
	timeService timeService,
	cfg config.AccessToken) *service {
	return &service{
		jwtGenerator: jwtGenerator,
		timeService:  timeService,
		cfg:          cfg,
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

	signedJWT, err := s.jwtGenerator.GenerateSignedJWT(claims, []byte(s.cfg.Secret))
	if err != nil {
		return nil, err
	}

	return &models.AccessToken{
		Token:     signedJWT,
		TokenType: tokenTypeBearer,
		ExpiresIn: int(s.cfg.TTL.Seconds()),
		ExpiresAt: expiresAt,
	}, nil
}
