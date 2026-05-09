package email_verification_tokens

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
)

type emailVerificationTokensConverter interface {
	ToEntity(token *models.EmailVerificationToken) (*Entity, error)
	ToModel(entity *Entity) *models.EmailVerificationToken
}

type repository struct {
	emailVerificationTokensConverter emailVerificationTokensConverter
}

func NewRepository(emailVerificationTokensConverter emailVerificationTokensConverter) *repository {
	return &repository{
		emailVerificationTokensConverter: emailVerificationTokensConverter,
	}
}

func (r *repository) Create(ctx context.Context, token *models.EmailVerificationToken) error {
	persist, err := database.FromCtx(ctx)
	if err != nil {
		return err
	}

	entity, err := r.emailVerificationTokensConverter.ToEntity(token)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO auth_service.email_verification_tokens (id, user_id, token_hash, expires_at, created_at)
			  VALUES (:id, :user_id, :token_hash, :expires_at, :created_at)`

	if _, err := persist.NamedExecContext(ctx, insertQuery, entity); err != nil {
		return err
	}

	return nil
}

func (r *repository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	persist, err := database.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, user_id, token_hash, expires_at, created_at FROM auth_service.email_verification_tokens 
			WHERE token_hash = $1 FOR UPDATE`

	var entity Entity
	if err := persist.GetContext(ctx, &entity, query, tokenHash); err != nil {
		return nil, err
	}

	return r.emailVerificationTokensConverter.ToModel(&entity), nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM auth_service.email_verification_tokens WHERE id = $1`

	return r.delete(ctx, query, id)
}

func (r *repository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM auth_service.email_verification_tokens WHERE user_id = $1`

	return r.delete(ctx, query, userID)
}

func (r *repository) delete(ctx context.Context, query string, key string) error {
	persist, err := database.FromCtx(ctx)
	if err != nil {
		return err
	}

	if _, err = persist.ExecContext(ctx, query, key); err != nil {
		return err
	}

	return nil
}
