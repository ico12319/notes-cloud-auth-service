package users

import "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (*converter) ToModel(entity Entity) *models.User {
	return &models.User{
		ID:    entity.ID,
		Name:  entity.Name,
		Email: entity.Email,
	}
}

func (c *converter) ManyToModel(entities []Entity) []*models.User {
	users := make([]*models.User, 0)

	for index := range entities {
		users = append(users, c.ToModel(entities[index]))
	}

	return users
}
