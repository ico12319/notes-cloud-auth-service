package refresh_tokens

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/api_errors"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"log"
)

type refreshTokenConverter interface {
	ToEntity(refreshToken *models.RefreshToken) (*Entity, error)
	ToModel(entity *Entity) *models.RefreshToken
}

type repository struct {
	refreshTokenConverter refreshTokenConverter
}

func NewRepository(refreshTokenConverter refreshTokenConverter) *repository {
	return &repository{
		refreshTokenConverter: refreshTokenConverter,
	}
}

func (r *repository) Create(ctx context.Context, refreshToken *models.RefreshToken) error {
	entity, err := r.refreshTokenConverter.ToEntity(refreshToken)
	if err != nil {
		return err
	}

	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO auth_service.refresh_tokens (id, user_id, token_hash, expires_at, revoked_at)
			  VALUES (:id, :user_id, :token_hash, :expires_at, :revoked_at)`

	if _, err := persistence.NamedExecContext(ctx, insertQuery, entity); err != nil {
		return err
	}

	return nil
}

func (r *repository) Revoke(ctx context.Context, id string) error {
	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return err
	}

	updateQuery := `UPDATE auth_service.refresh_tokens SET revoked_at = now() WHERE id = $1`

	result, err := persistence.ExecContext(ctx, updateQuery, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("failed to check how many rows were affected when revoking refresh token with id %s", id)

		return err
	}

	if rowsAffected == 0 {
		return api_errors.ErrTokenNotFound
	}

	return nil
}

func (r *repository) Get(ctx context.Context, id string) (*models.RefreshToken, error) {
	persistence, err := database.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	selectQuery := `SELECT id, user_id, token_hash, expires_at, revoked_at, created_at FROM auth_service.refresh_tokens 
					WHERE id = $1 FOR UPDATE`

	var entity Entity
	if err := persistence.GetContext(ctx, &entity, selectQuery, id); err != nil {
		return nil, err
	}

	return r.refreshTokenConverter.ToModel(&entity), nil
}
