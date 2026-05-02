package refresh_tokens

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"log"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (*converter) ToEntity(refreshToken *models.RefreshToken, tokenHash string) (*Entity, error) {
	stringifiedID, err := uuid.Parse(refreshToken.ID)
	if err != nil {
		log.Printf("failed to stringify refresh token id %s: err: %s", refreshToken.ID, err.Error())

		return nil, err
	}

	stringifiedUserID, err := uuid.Parse(refreshToken.UserID)
	if err != nil {
		log.Printf("failed to stringify user id %s in refresh token with id %s: err: %s", refreshToken.UserID,
			refreshToken.ID, err.Error())

		return nil, err
	}

	entity := &Entity{
		ID:        stringifiedID,
		UserID:    stringifiedUserID,
		TokenHash: tokenHash,
		ExpiresAt: refreshToken.ExpiresAt,
		CreatedAt: refreshToken.CreatedAt,
	}

	if refreshToken.RevokedAt != nil {
		entity.RevokedAt = sql.NullTime{
			Time:  *refreshToken.RevokedAt,
			Valid: true,
		}
	}

	return entity, nil
}

func (*converter) ToModel(entity *Entity) *models.RefreshToken {
	refreshToken := &models.RefreshToken{
		ID:        entity.ID.String(),
		UserID:    entity.UserID.String(),
		ExpiresAt: entity.ExpiresAt,
		CreatedAt: entity.CreatedAt,
	}

	if entity.RevokedAt.Valid {
		refreshToken.RevokedAt = &entity.RevokedAt.Time
	}

	return refreshToken
}
