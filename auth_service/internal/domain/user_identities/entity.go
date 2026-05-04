package user_identities

import (
	"github.com/google/uuid"
	"time"
)

type Entity struct {
	ID             uuid.UUID `db:"id"`
	UserID         uuid.UUID `db:"user_id"`
	Provider       string    `db:"provider"`
	ProviderUserID string    `db:"provider_user_id"`
	Email          string    `db:"email"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
