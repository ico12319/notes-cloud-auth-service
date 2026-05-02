package users

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Entity struct {
	ID           uuid.UUID    `db:"id"`
	FirstName    string       `db:"first_name"`
	LastName     string       `db:"last_name"`
	Email        string       `db:"email"`
	PasswordHash string       `db:"password_hash"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}
