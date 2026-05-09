package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/util"
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
	Update(ctx context.Context, id string, patch *models.UserPatch) (*models.User, error)
}

type identityService interface {
	Create(ctx context.Context, identity *models.UserIdentity) (*models.UserIdentity, error)
	Find(ctx context.Context, providerUserID, provider string) (*models.UserIdentity, error)
	UpdateEmail(ctx context.Context, id string, email string) (*models.UserIdentity, error)
}

type emailVerificationService interface {
	FindToken(ctx context.Context, token string) (*models.EmailVerificationToken, error)
	Delete(ctx context.Context, id string) error
}

var ErrInvalidVerificationCode = errors.New("invalid verification code")

type service struct {
	uuidService              uuidService
	passwordService          passwordService
	timeService              timeService
	userRepository           userRepository
	identityService          identityService
	emailVerificationService emailVerificationService
}

func NewService(
	userRepository userRepository,
	passwordService passwordService,
	timeService timeService,
	uuidService uuidService,
	identityService identityService,
	emailVerificationService emailVerificationService,
) *service {
	return &service{
		timeService:              timeService,
		passwordService:          passwordService,
		userRepository:           userRepository,
		uuidService:              uuidService,
		identityService:          identityService,
		emailVerificationService: emailVerificationService,
	}
}

func (s *service) Create(ctx context.Context, user *models.User, password *string) (*models.User, error) {
	registeredUser := &models.User{
		ID:            s.uuidService.Generate(),
		Name:          user.Name,
		Email:         user.Email,
		CreatedAt:     s.timeService.Now(),
		EmailVerified: user.EmailVerified,
	}

	if password != nil {
		log.Printf("password is not nil when registering, user %s is registering using non oauth flow", user.Email)

		passwordHash, err := s.passwordService.GeneratePasswordHash(*password)
		if err != nil {
			log.Printf("failed to generate password hash for user with email %s", user.Email)

			return nil, err
		}

		stringifiedPasswordHash := string(passwordHash)
		registeredUser.PasswordHash = &stringifiedPasswordHash
	}

	if err := s.userRepository.Create(ctx, registeredUser); err != nil {
		if database.IsUniqueViolation(err) {
			log.Printf("user with email %s already exist", user.Email)

			return nil, api_errors.ErrEmailAlreadyExist
		}

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

func (s *service) FindOrCreateByOAuthIdentity(ctx context.Context, userAuthInfo *models.UserAuthInfo) (*models.User, error) {
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

	createdUser, err := s.Create(ctx, &models.User{
		EmailVerified: true,
		Name:          userAuthInfo.Name,
		Email:         userAuthInfo.Email,
	}, nil)
	if err != nil {
		return nil, err
	}

	if _, err := s.identityService.Create(ctx, &models.UserIdentity{
		UserID:         createdUser.ID,
		Provider:       userAuthInfo.Provider,
		ProviderUserID: userAuthInfo.ProviderUserID,
		Email:          userAuthInfo.Email,
	}); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *service) UpdateUser(ctx context.Context, id string, patch *models.UserPatch) (*models.User, error) {
	return s.userRepository.Update(ctx, id, patch)
}

func (s *service) VerifyUser(ctx context.Context, verificationCode string) (*models.User, error) {
	emailVerificationToken, err := s.emailVerificationService.FindToken(ctx, verificationCode)
	if err != nil {
		return nil, ErrInvalidVerificationCode
	}

	if err := s.emailVerificationService.Delete(ctx, emailVerificationToken.ID); err != nil {
		return nil, err
	}

	user, err := s.UpdateUser(ctx, emailVerificationToken.UserID, &models.UserPatch{
		EmailVerified: util.GenericPtr(true),
	})
	if err != nil {
		return nil, err
	}

	return user, err
}
