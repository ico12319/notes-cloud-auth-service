package users

import (
	"context"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"strings"
	"time"
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

	insertQuery := `INSERT INTO auth_service.users (id, display_name, email, password_hash, created_at, email_verified)
			  VALUES (:id, :display_name, :email, :password_hash, :created_at, :email_verified)`

	if _, err = persistence.NamedExecContext(ctx, insertQuery, entity); err != nil {
		return err
	}

	return nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	selectQuery := `SELECT id, display_name, email, password_hash, created_at, updated_at, email_verified                                                                                                                                                 
                  FROM auth_service.users WHERE email = $1`

	return r.get(ctx, selectQuery, email)
}

func (r *repository) GetByID(ctx context.Context, id string) (*models.User, error) {
	selectQuery := `SELECT id, display_name, email, password_hash, created_at, updated_at, email_verified                                                                                                                                                      
                  FROM auth_service.users WHERE id = $1`

	return r.get(ctx, selectQuery, id)
}

func (r *repository) Update(ctx context.Context, id string, patch *models.UserPatch) (*models.User, error) {
	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var setClauses []string
	args := map[string]interface{}{
		"id":         id,
		"updated_at": time.Now(),
	}

	if patch.DisplayName != nil {
		setClauses = append(setClauses, "display_name = :display_name")
		args["display_name"] = *patch.DisplayName
	}

	if patch.Email != nil {
		setClauses = append(setClauses, "email = :email")
		args["email"] = *patch.Email
	}

	if patch.EmailVerified != nil {
		setClauses = append(setClauses, "email_verified = :email_verified")
		args["email_verified"] = *patch.EmailVerified
	}

	if patch.PasswordHash != nil {
		setClauses = append(setClauses, "password_hash = :password_hash")
		args["password_hash"] = *patch.PasswordHash
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = :updated_at")

	updateQuery := fmt.Sprintf(
		"UPDATE auth_service.users SET %s WHERE id = :id",
		strings.Join(setClauses, ", "),
	)

	if _, err = persistence.NamedExecContext(ctx, updateQuery, args); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
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
