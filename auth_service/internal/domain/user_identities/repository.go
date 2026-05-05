package user_identities

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"time"
)

type identityConverter interface {
	ToEntity(identity *models.UserIdentity) (*Entity, error)
	ToModel(entity *Entity) *models.UserIdentity
}

type repository struct {
	identityConverter identityConverter
}

func NewRepository(identityConverter identityConverter) *repository {
	return &repository{identityConverter: identityConverter}
}

func (r *repository) Create(ctx context.Context, identity *models.UserIdentity) error {
	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return err
	}

	entity, err := r.identityConverter.ToEntity(identity)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO auth_service.identities (id, user_id, provider, provider_user_id, email, created_at)
			  VALUES (:id, :user_id, :provider, :provider_user_id, :email, :created_at)`

	if _, err = persistence.NamedExecContext(ctx, insertQuery, entity); err != nil {
		return err
	}

	return nil
}

func (r *repository) Get(ctx context.Context, providerUserID, provider string) (*models.UserIdentity, error) {
	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	selectQuery := `SELECT id, user_id, provider, provider_user_id, email, created_at, updated_at FROM auth_service.identities WHERE provider_user_id = $1 AND provider = $2`

	var entity Entity
	if err := persistence.GetContext(ctx, &entity, selectQuery, providerUserID, provider); err != nil {
		return nil, err
	}

	return r.identityConverter.ToModel(&entity), nil
}

func (r *repository) Update(ctx context.Context, id string, email string, updatedAt time.Time) error {
	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return err
	}

	updateQuery := `UPDATE auth_service.identities SET email = $1, updated_at = $2 WHERE id = $3`

	_, err = persistence.ExecContext(ctx, updateQuery, email, updatedAt, id)
	return err
}
