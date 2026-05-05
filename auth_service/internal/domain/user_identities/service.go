package user_identities

import (
	"context"
	"database/sql"
	"errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"time"
)

type userIdentityRepository interface {
	Create(ctx context.Context, identity *models.UserIdentity) error
	Get(ctx context.Context, providerUserID, provider string) (*models.UserIdentity, error)
	Update(ctx context.Context, id string, email string, updatedAt time.Time) error
}

type uuidService interface {
	Generate() string
}

type timeService interface {
	Now() time.Time
}

type service struct {
	userIdentityRepository userIdentityRepository
	uuidService            uuidService
	timeService            timeService
}

func NewService(
	userIdentityRepository userIdentityRepository,
	uuidService uuidService,
	timeService timeService,
) *service {
	return &service{
		userIdentityRepository: userIdentityRepository,
		uuidService:            uuidService,
		timeService:            timeService,
	}
}

func (s *service) Create(ctx context.Context, identity *models.UserIdentity) (*models.UserIdentity, error) {
	identity.ID = s.uuidService.Generate()
	identity.CreatedAt = s.timeService.Now()

	if err := s.userIdentityRepository.Create(ctx, identity); err != nil {
		return nil, err
	}

	return identity, nil
}

func (s *service) Find(ctx context.Context, providerUserID, provider string) (*models.UserIdentity, error) {
	identity, err := s.userIdentityRepository.Get(ctx, providerUserID, provider)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api_errors.ErrIdentityNotFound
		}

		return nil, err
	}

	return identity, nil
}

func (s *service) UpdateEmail(ctx context.Context, identityID string, email string) (*models.UserIdentity, error) {
	if err := s.userIdentityRepository.Update(ctx, identityID, email, s.timeService.Now()); err != nil {
		return nil, err
	}
	return nil, nil
}
