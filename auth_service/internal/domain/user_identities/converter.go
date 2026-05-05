package user_identities

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/util"
	"log"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (*converter) ToEntity(identity *models.UserIdentity) (*Entity, error) {
	stringifiedID, err := uuid.Parse(identity.ID)
	if err != nil {
		log.Printf("failed to stringify identity id %s: err: %s", identity.ID, err.Error())

		return nil, err
	}

	stringifiedUserID, err := uuid.Parse(identity.UserID)
	if err != nil {
		log.Printf("failed to stringify identity user_id %s: err: %s", identity.UserID, err.Error())

		return nil, err
	}

	var updatedAt sql.NullTime
	if identity.UpdatedAt != nil {
		updatedAt = sql.NullTime{Time: *identity.UpdatedAt, Valid: true}
	}

	return &Entity{
		ID:             stringifiedID,
		UserID:         stringifiedUserID,
		Provider:       identity.Provider,
		ProviderUserID: identity.ProviderUserID,
		Email:          identity.Email,
		CreatedAt:      identity.CreatedAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func (*converter) ToModel(entity *Entity) *models.UserIdentity {
	return &models.UserIdentity{
		ID:             entity.ID.String(),
		UserID:         entity.UserID.String(),
		Email:          entity.Email,
		Provider:       entity.Provider,
		ProviderUserID: entity.ProviderUserID,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      util.NullTimeToPtr(entity.UpdatedAt),
	}
}
