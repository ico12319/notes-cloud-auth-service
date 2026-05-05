package users

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Entity struct {
	ID           uuid.UUID    `db:"id"`
	DisplayName  string       `db:"display_name"`
	Email        string       `db:"email"`
	PasswordHash *string      `db:"password_hash"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}
