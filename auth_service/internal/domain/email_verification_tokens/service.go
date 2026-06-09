package email_verification_tokens

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"time"
)

type emailVerificationRepository interface {
	Create(ctx context.Context, token *models.EmailVerificationToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type uuidGenerator interface {
	Generate() string
}

type timeGenerator interface {
	Now() time.Time
}

type tokenHasher interface {
	Hash(rawToken string) string
}

type service struct {
	emailVerificationRepository emailVerificationRepository
	uuidGenerator               uuidGenerator
	timeGenerator               timeGenerator
	tokenHasher                 tokenHasher
}

func NewService(
	emailVerificationRepository emailVerificationRepository,
	uuidGenerator uuidGenerator,
	timeGenerator timeGenerator,
	tokenHasher tokenHasher,
) *service {
	return &service{
		emailVerificationRepository: emailVerificationRepository,
		uuidGenerator:               uuidGenerator,
		timeGenerator:               timeGenerator,
		tokenHasher:                 tokenHasher,
	}
}

func (s *service) Create(ctx context.Context, verificationCode, userID string) (*models.EmailVerificationToken, error) {
	emailVerificationCode := &models.EmailVerificationToken{
		ID:        s.uuidGenerator.Generate(),
		UserID:    userID,
		TokenHash: s.tokenHasher.Hash(verificationCode),
		ExpiresAt: s.timeGenerator.Now().Add(180 * time.Second),
		CreatedAt: s.timeGenerator.Now(),
	}

	if err := s.emailVerificationRepository.Create(ctx, emailVerificationCode); err != nil {
		return nil, err
	}

	return emailVerificationCode, nil
}

func (s *service) FindToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	return s.emailVerificationRepository.GetByTokenHash(ctx, s.tokenHasher.Hash(token))
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.emailVerificationRepository.Delete(ctx, id)
}

func (s *service) DeleteByUserID(ctx context.Context, userID string) error {
	return s.emailVerificationRepository.DeleteByUserID(ctx, userID)
}
