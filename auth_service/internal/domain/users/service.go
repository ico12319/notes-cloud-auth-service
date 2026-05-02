package users

import (
	"context"
	"fmt"
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

type service struct {
	uuidService     uuidService
	passwordService passwordService
	timeService     timeService
	userRepository  userRepository
}

func NewService(
	userRepository userRepository,
	passwordService passwordService,
	timeService timeService,
	uuidService uuidService,
) *service {
	return &service{
		timeService:     timeService,
		passwordService: passwordService,
		userRepository:  userRepository,
		uuidService:     uuidService,
	}
}

func (s *service) Create(ctx context.Context, registerRequest *request_models.RegisterRequest) (*models.User, error) {
	if registerRequest == nil {
		return nil, fmt.Errorf("register input can't be nil")
	}

	passwordHash, err := s.passwordService.GeneratePasswordHash(registerRequest.Password)
	if err != nil {
		log.Printf("failed to generate password hash for user with email %s", registerRequest.Email)

		return nil, err
	}

	registeredUser := &models.User{
		ID:           s.uuidService.Generate(),
		FirstName:    registerRequest.FirstName,
		LastName:     registerRequest.LastName,
		Email:        registerRequest.Email,
		CreatedAt:    s.timeService.Now(),
		PasswordHash: string(passwordHash),
	}

	if err := s.userRepository.Create(ctx, registeredUser); err != nil {
		log.Printf("kude gurmish be da ta eba %s", err.Error())
		return nil, err
	}

	return registeredUser, nil
}

func (s *service) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepository.GetByEmail(ctx, email)
}

func (s *service) FindByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepository.GetByID(ctx, id)
}
