package users

import (
	"github.com/google/uuid"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/util"
	"log"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (*converter) ToEntity(user *models.User) (*Entity, error) {
	stringifiedID, err := uuid.Parse(user.ID)
	if err != nil {
		log.Printf("failed to stringify user id %s: err: %s", user.ID, err.Error())

		return nil, err
	}

	return &Entity{
		ID:            stringifiedID,
		DisplayName:   user.Name,
		Email:         user.Email,
		PasswordHash:  user.PasswordHash,
		CreatedAt:     user.CreatedAt,
		EmailVerified: user.EmailVerified,
	}, nil
}

func (*converter) ToModel(entity *Entity) *models.User {
	return &models.User{
		ID:            entity.ID.String(),
		Name:          entity.DisplayName,
		Email:         entity.Email,
		CreatedAt:     entity.CreatedAt,
		PasswordHash:  entity.PasswordHash,
		UpdatedAt:     util.NullTimeToPtr(entity.UpdatedAt),
		EmailVerified: entity.EmailVerified,
	}
}

func (*converter) ToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:            user.ID,
		DisplayName:   user.Name,
		CreatedAt:     user.CreatedAt,
		Email:         user.Email,
		UpdatedAt:     user.UpdatedAt,
		EmailVerified: user.EmailVerified,
	}
}
