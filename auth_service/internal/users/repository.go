package users

import (
	"context"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/database"
)

type repository struct{}

func NewRepository() *repository {
	return &repository{}
}

func (r *repository) GetAll(ctx context.Context) ([]Entity, error) {
	tx, err := database.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var entities []Entity
	if err := tx.SelectContext(ctx, &entities, `SELECT id, name, email FROM users`); err != nil {
		return nil, err
	}

	return entities, nil
}
