package auth

import (
	"context"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"log"
)

type userService interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type accessTokenService interface {
	GenerateForUser(userID string) (*models.AccessToken, error)
}

type refreshTokenService interface {
	GenerateForUser(ctx context.Context, userID string) (*models.RefreshToken, error)
}

type passwordService interface {
	CompareHashAndPassword(hashedPassword string, password string) error
}

var ErrWrongLoginCredentials = fmt.Errorf("wrong email or password provided")

type service struct {
	userService         userService
	accessTokenService  accessTokenService
	refreshTokenService refreshTokenService
	passwordService     passwordService
}

func NewService(
	userService userService,
	accessTokenService accessTokenService,
	refreshTokenService refreshTokenService,
	passwordService passwordService,
) *service {
	return &service{
		userService:         userService,
		refreshTokenService: refreshTokenService,
		accessTokenService:  accessTokenService,
		passwordService:     passwordService,
	}
}

func (s *service) Login(ctx context.Context, request *request_models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userService.FindByEmail(ctx, request.Email)
	if err != nil {
		log.Printf("failed to find user with email %s", request.Email)

		return nil, ErrWrongLoginCredentials
	}

	if err := s.passwordService.CompareHashAndPassword(user.PasswordHash, request.Password); err != nil {
		return nil, ErrWrongLoginCredentials
	}

	accessToken, err := s.accessTokenService.GenerateForUser(user.ID)
	if err != nil {
		log.Printf("failed to generate access token for user with id %s", user.ID)

		return nil, fmt.Errorf("failed to generate access token for user %w", err)
	}

	refreshToken, err := s.refreshTokenService.GenerateForUser(ctx, user.ID)
	if err != nil {
		log.Printf("failed to generate refresh token for user with id %s", user.ID)

		return nil, fmt.Errorf("failed to generate refresh token for user %w", err)
	}

	return &models.LoginResponse{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.RawToken,
		TokenType:    accessToken.TokenType,
	}, nil
}
