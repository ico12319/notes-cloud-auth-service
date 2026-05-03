package auth

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"log"
)

type userService interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type tokenBundleService interface {
	GenerateBundle(ctx context.Context, userID string) (*models.TokenBundle, error)
}

type passwordService interface {
	CompareHashAndPassword(hashedPassword string, password string) error
}

type service struct {
	userService        userService
	tokenBundleService tokenBundleService
	passwordService    passwordService
}

func NewService(
	userService userService,
	tokenBundleService tokenBundleService,
	passwordService passwordService,
) *service {
	return &service{
		userService:        userService,
		tokenBundleService: tokenBundleService,
		passwordService:    passwordService,
	}
}

func (s *service) Login(ctx context.Context, request *request_models.LoginRequest) (*models.TokenBundle, error) {
	user, err := s.userService.FindByEmail(ctx, request.Email)
	if err != nil {
		log.Printf("failed to find user with email %s", request.Email)

		return nil, api_errors.ErrWrongLoginCredentials
	}

	if err := s.passwordService.CompareHashAndPassword(user.PasswordHash, request.Password); err != nil {
		return nil, api_errors.ErrWrongLoginCredentials
	}

	return s.tokenBundleService.GenerateBundle(ctx, user.ID)
}
