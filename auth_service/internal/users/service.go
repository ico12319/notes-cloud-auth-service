package users

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
)

type userRepository interface {
	GetAll(ctx context.Context) ([]Entity, error)
}

type userConverter interface {
	ManyToModel(entities []Entity) []*models.User
}

type service struct {
	userRepository userRepository
	userConverter  userConverter
}

func NewService(
	userRepository userRepository,
	userConverter userConverter) *service {
	return &service{
		userRepository: userRepository,
		userConverter:  userConverter,
	}
}

func (s *service) GetAll(ctx context.Context) ([]*models.User, error) {
	entities, err := s.userRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return s.userConverter.ManyToModel(entities), nil
}
