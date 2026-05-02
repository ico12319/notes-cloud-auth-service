package refresh_tokens

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"log"
	"strings"
	"time"
)

const (
	refreshTokenPrefix      = "rfr"
	refreshTokenSecretBytes = 32
)

type uuidService interface {
	Generate() string
}

type stringEncoder interface {
	EncodeToString(secret []byte) string
}

type randomService interface {
	Read(b []byte) (n int, err error)
}

type refreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *models.RefreshToken, tokenHash string) error
	Revoke(ctx context.Context, tokenID string) error
}

type timeService interface {
	Now() time.Time
}

type service struct {
	timeService            timeService
	uuidService            uuidService
	stringEncoder          stringEncoder
	randomService          randomService
	refreshTokenRepository refreshTokenRepository
	refreshTokenSecret     string
}

func NewService(
	refreshTokenRepository refreshTokenRepository,
	randomService randomService,
	stringEncoder stringEncoder,
	uuidService uuidService,
	timeService timeService,
	refreshTokenSecret string,
) *service {
	return &service{
		refreshTokenRepository: refreshTokenRepository,
		stringEncoder:          stringEncoder,
		randomService:          randomService,
		uuidService:            uuidService,
		timeService:            timeService,
		refreshTokenSecret:     refreshTokenSecret,
	}
}

func (s *service) GenerateForUser(ctx context.Context, userID string) (*models.RefreshToken, error) {
	refreshTokenID := s.uuidService.Generate()

	secretBytes := make([]byte, refreshTokenSecretBytes)
	if _, err := s.randomService.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("generate refresh token secret: %w", err)
	}

	secret := s.stringEncoder.EncodeToString(secretBytes)
	rawToken := fmt.Sprintf("%s_%s.%s", refreshTokenPrefix, refreshTokenID, secret)

	generatedRefreshToken := &models.RefreshToken{
		ID:        refreshTokenID,
		UserID:    userID,
		RawToken:  rawToken,
		ExpiresAt: s.timeService.Now().Add(30 * 24 * time.Hour),
	}

	if err := s.refreshTokenRepository.Create(ctx, generatedRefreshToken, s.hashRefreshToken(rawToken, []byte(s.refreshTokenSecret))); err != nil {
		return nil, err
	}

	return generatedRefreshToken, nil
}

func (s *service) Revoke(ctx context.Context, rawToken string) error {
	tokenID, err := s.extractIDFromRawToken(rawToken)
	if err != nil {
		log.Printf("failed to extract id from raw token %s", err.Error())

		return err
	}

	if err := s.refreshTokenRepository.Revoke(ctx, tokenID); err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			log.Printf("refresh token with id %s is already deleted, won't revoked it", tokenID)

			return nil
		}

		return err
	}

	return nil
}

func (s *service) hashRefreshToken(rawToken string, secretKey []byte) string {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(rawToken))

	return s.stringEncoder.EncodeToString(mac.Sum(nil))
}

func (*service) extractIDFromRawToken(rawToken string) (string, error) {
	prefix := refreshTokenPrefix + "_"

	if !strings.HasPrefix(rawToken, prefix) {
		return "", fmt.Errorf("invalid refresh token prefix")
	}

	parts := strings.Split(rawToken, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid refresh token format")
	}

	rawID := strings.TrimPrefix(parts[0], prefix)
	if rawID == "" {
		return "", fmt.Errorf("missing refresh token id")
	}

	if _, err := uuid.Parse(rawID); err != nil {
		return "", fmt.Errorf("invalid refresh token id %q: %w", rawID, err)
	}

	return rawID, nil
}
