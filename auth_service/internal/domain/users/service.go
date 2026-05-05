package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/request_models"
	"log"
	"time"
)

type passwordService interface {
	GeneratePasswordHash(password string) ([]byte, error)
}

type uuidService interface {
	Generate() string
}

type timeService interface {
	Now() time.Time
}

type userRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type identityService interface {
	Create(ctx context.Context, identity *models.UserIdentity) (*models.UserIdentity, error)
	Find(ctx context.Context, providerUserID, provider string) (*models.UserIdentity, error)
	UpdateEmail(ctx context.Context, id string, email string) (*models.UserIdentity, error)
}

type service struct {
	uuidService     uuidService
	passwordService passwordService
	timeService     timeService
	userRepository  userRepository
	identityService identityService
}

func NewService(
	userRepository userRepository,
	passwordService passwordService,
	timeService timeService,
	uuidService uuidService,
	identityService identityService,
) *service {
	return &service{
		timeService:     timeService,
		passwordService: passwordService,
		userRepository:  userRepository,
		uuidService:     uuidService,
		identityService: identityService,
	}
}

func (s *service) Create(ctx context.Context, registerRequest *request_models.RegisterRequest) (*models.User, error) {
	if registerRequest == nil {
		return nil, fmt.Errorf("register input can't be nil")
	}

	registeredUser := &models.User{
		ID:        s.uuidService.Generate(),
		FirstName: registerRequest.FirstName,
		LastName:  registerRequest.LastName,
		Email:     registerRequest.Email,
		CreatedAt: s.timeService.Now(),
	}

	if registerRequest.Password != nil {
		log.Printf("password is not nil when registering, user %s is registering using non oauth flow", registerRequest.Email)

		passwordHash, err := s.passwordService.GeneratePasswordHash(*registerRequest.Password)
		if err != nil {
			log.Printf("failed to generate password hash for user with email %s", registerRequest.Email)

			return nil, err
		}

		stringifiedPasswordHash := string(passwordHash)
		registeredUser.PasswordHash = &stringifiedPasswordHash
	}

	if err := s.userRepository.Create(ctx, registeredUser); err != nil {
		return nil, err
	}

	return registeredUser, nil
}

func (s *service) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api_errors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *service) FindByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepository.GetByID(ctx, id)
}

func (s *service) ResolveUser(ctx context.Context, userAuthInfo *models.UserAuthInfo) (*models.User, error) {
	identity, err := s.identityService.Find(ctx, userAuthInfo.ProviderUserID, userAuthInfo.Provider)
	if err != nil && !api_errors.IsIdentityNotFoundError(err) {
		return nil, err
	}

	if identity != nil {
		if identity.Email != userAuthInfo.Email {
			if _, err := s.identityService.UpdateEmail(ctx, identity.UserID, userAuthInfo.Email); err != nil {
				return nil, err
			}
		}

		return s.FindByID(ctx, identity.UserID)
	}

	log.Printf("identity from provider %s with provider id %s does not exist",
		userAuthInfo.Provider, userAuthInfo.ProviderUserID)

	userID, err := s.resolveOrCreateUser(ctx, userAuthInfo)
	if err != nil {
		return nil, err
	}

	if _, err := s.identityService.Create(ctx, &models.UserIdentity{
		UserID:         userID,
		Provider:       userAuthInfo.Provider,
		ProviderUserID: userAuthInfo.ProviderUserID,
		Email:          userAuthInfo.Email,
	}); err != nil {
		return nil, err
	}

	return s.FindByID(ctx, userID)
}

func (s *service) resolveOrCreateUser(ctx context.Context, userAuthInfo *models.UserAuthInfo) (string, error) {
	existingUser, err := s.FindByEmail(ctx, userAuthInfo.Email)
	if err != nil && !api_errors.IsUserNotFoundError(err) {
		return "", err
	}

	if existingUser != nil {
		return existingUser.ID, nil
	}

	createdUser, err := s.Create(ctx, &request_models.RegisterRequest{
		Email:     userAuthInfo.Email,
		FirstName: userAuthInfo.FirstName,
		LastName:  userAuthInfo.LastName,
	})
	if err != nil {
		return "", err
	}

	return createdUser.ID, nil
}
