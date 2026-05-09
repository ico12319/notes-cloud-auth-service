package email_verification_tokens

import (
	"github.com/google/uuid"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (*converter) ToEntity(code *models.EmailVerificationToken) (*Entity, error) {
	stringifiedID, err := uuid.Parse(code.ID)
	if err != nil {
		return nil, err
	}
	stringifiedUserID, err := uuid.Parse(code.UserID)
	if err != nil {
		return nil, err
	}

	return &Entity{
		ID:        stringifiedID,
		UserID:    stringifiedUserID,
		TokenHash: code.TokenHash,
		ExpiresAt: code.ExpiresAt,
		CreatedAt: code.CreatedAt,
	}, nil
}

func (*converter) ToModel(entity *Entity) *models.EmailVerificationToken {
	return &models.EmailVerificationToken{
		ID:        entity.ID.String(),
		UserID:    entity.UserID.String(),
		TokenHash: entity.TokenHash,
		ExpiresAt: entity.ExpiresAt,
		CreatedAt: entity.CreatedAt,
	}
}
