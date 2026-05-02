package users

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"log"
)

type userConverter interface {
	ToEntity(user *models.User, passwordHash []byte) (*Entity, error)
}

type repository struct {
	userConverter userConverter
}

func NewRepository(userConverter userConverter) *repository {
	return &repository{
		userConverter: userConverter,
	}
}

func (r *repository) Create(ctx context.Context, passwordHash []byte, user *models.User) error {
	entity, err := r.userConverter.ToEntity(user, passwordHash)
	if err != nil {
		log.Printf("nqma da e tuka ama vse pak %s", err.Error())
		return err
	}

	persistence, err := database.FromCtx(ctx)
	if err != nil {
		log.Printf("i tuka nqma ama da vidim %s", err.Error())
		return err
	}

	insertQuery := `INSERT INTO auth_service.users (id, first_name, last_name, email, password_hash, created_at)
			  VALUES (:id, :first_name, :last_name, :email, :password_hash, :created_at)`

	if _, err = persistence.NamedExecContext(ctx, insertQuery, entity); err != nil {
		log.Printf("ei tuka smurdi ama da vidim %s", err.Error())
		return err
	}

	return nil
}
