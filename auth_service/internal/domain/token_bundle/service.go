package token_bundle

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
	"log"
	"time"
)

type accessTokenGenerator interface {
	GenerateForUser(userID string) (*accesstoken.Token, error)
}

type refreshTokenService interface {
	Get(ctx context.Context, rawToken string) (*models.RefreshToken, error)
	GenerateForUser(ctx context.Context, userID string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, rawToken string) error
}

type userService interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
}

type timeService interface {
	Now() time.Time
}

type service struct {
	accessTokenGenerator accessTokenGenerator
	refreshTokenService  refreshTokenService
	userService          userService
	timeService          timeService
}

func NewService(
	accessTokenGenerator accessTokenGenerator,
	refreshTokenService refreshTokenService,
	userService userService,
	timeService timeService,
) *service {
	return &service{
		accessTokenGenerator: accessTokenGenerator,
		refreshTokenService:  refreshTokenService,
		userService:          userService,
		timeService:          timeService,
	}
}

func (s *service) GenerateBundle(ctx context.Context, userID string) (*models.TokenBundle, error) {
	accessToken, err := s.accessTokenGenerator.GenerateForUser(userID)
	if err != nil {
		log.Printf("failed to generate a fres new access token for user with id %s", userID)

		return nil, err
	}

	refreshedRefreshToken, err := s.refreshTokenService.GenerateForUser(ctx, userID)
	if err != nil {
		log.Printf("failed to generate a fres new refresh token for user with id %s", userID)

		return nil, err
	}

	return &models.TokenBundle{
		RefreshToken: refreshedRefreshToken.RawToken,
		AccessToken:  accessToken,
	}, nil
}

func (s *service) Refresh(ctx context.Context, request *request_models.RefreshRequest) (*models.TokenBundle, error) {
	refreshToken, err := s.refreshTokenService.Get(ctx, request.RefreshToken)
	if err != nil {
		return nil, err
	}

	if _, err := s.userService.FindByID(ctx, refreshToken.UserID); err != nil {
		log.Printf("failed to find user with %s when trying to refresh token bundle", refreshToken.UserID)

		return nil, err
	}

	if refreshToken.RevokedAt != nil {
		log.Printf("refresh token with id %s for user %s is already revoked on %v, won't generate fresh token bundle", refreshToken.ID,
			refreshToken.RevokedAt, refreshToken.UserID)

		return nil, api_errors.ErrRefreshTokenRevoked
	}

	if s.timeService.Now().UTC().After(refreshToken.ExpiresAt) {
		log.Printf("refresh token with id %s for user %s is expired, won't generate fresh token bundle", refreshToken.ID,
			refreshToken.UserID)

		return nil, api_errors.ErrExpiredRefreshToken
	}

	if err := s.refreshTokenService.Revoke(ctx, request.RefreshToken); err != nil {
		log.Printf("failed to revoke refresh token with id %s for user with id %s", refreshToken.ID,
			refreshToken.UserID)

		return nil, err
	}

	return s.GenerateBundle(ctx, refreshToken.UserID)
}
