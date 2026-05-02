package users

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
)

type userConverter interface {
	ToEntity(user *models.User) (*Entity, error)
	ToModel(entity *Entity) *models.User
}

type repository struct {
	userConverter userConverter
}

func NewRepository(userConverter userConverter) *repository {
	return &repository{
		userConverter: userConverter,
	}
}

func (r *repository) Create(ctx context.Context, user *models.User) error {
	entity, err := r.userConverter.ToEntity(user)
	if err != nil {
		return err
	}

	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO auth_service.users (id, first_name, last_name, email, password_hash, created_at)
			  VALUES (:id, :first_name, :last_name, :email, :password_hash, :created_at)`

	if _, err = persistence.NamedExecContext(ctx, insertQuery, entity); err != nil {
		return err
	}

	return nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	selectQuery := `SELECT id, first_name, last_name, email, password_hash, created_at, updated_at                                                                                                                                                 
                  FROM auth_service.users WHERE email = $1`

	return r.get(ctx, selectQuery, email)
}

func (r *repository) GetByID(ctx context.Context, id string) (*models.User, error) {
	selectQuery := `SELECT id, first_name, last_name, email, password_hash, created_at, updated_at                                                                                                                                                 
                  FROM auth_service.users WHERE id = $1`

	return r.get(ctx, selectQuery, id)
}

func (r *repository) get(ctx context.Context, query string, key string) (*models.User, error) {
	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var entity Entity
	if err := persistence.GetContext(ctx, &entity, query, key); err != nil {
		return nil, err
	}

	return r.userConverter.ToModel(&entity), nil
}
