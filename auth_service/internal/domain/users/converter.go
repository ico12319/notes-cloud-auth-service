package users

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (*converter) ToEntity(user *models.User, passwordHash []byte) (*Entity, error) {
	stringifiedID, err := uuid.Parse(user.ID)
	if err != nil {
		log.Printf("failed to stringify user id %s: err: %s", user.ID, err.Error())

		return nil, err
	}

	return &Entity{
		ID:           stringifiedID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		PasswordHash: string(passwordHash),
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (*converter) ToModel(entity *Entity) *models.User {
	return &models.User{
		ID:        entity.ID.String(),
		FirstName: entity.FirstName,
		LastName:  entity.LastName,
		Email:     entity.Email,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: nullTimeToPtr(entity.UpdatedAt),
	}
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
