package database

import (
	"errors"

	"github.com/lib/pq"
)

// PostgreSQL error codes
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	UniqueViolation     pq.ErrorCode = "23505"
	ForeignKeyViolation pq.ErrorCode = "23503"
	NotNullViolation    pq.ErrorCode = "23502"
	CheckViolation      pq.ErrorCode = "23514"
)

var (
	ErrUniqueViolation     = errors.New("record already exists")
	ErrForeignKeyViolation = errors.New("referenced record not found")
	ErrNotNullViolation    = errors.New("required field is missing")
	ErrCheckViolation      = errors.New("value violates check constraint")
)

func MapError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case UniqueViolation:
			return ErrUniqueViolation
		case ForeignKeyViolation:
			return ErrForeignKeyViolation
		case NotNullViolation:
			return ErrNotNullViolation
		case CheckViolation:
			return ErrCheckViolation
		}
	}

	return err
}

// ConstraintName returns the constraint name from a postgres error, if available
func ConstraintName(err error) string {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Constraint
	}
	return ""
}

// IsUniqueViolation checks if the error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == UniqueViolation
	}
	return false
}

// IsForeignKeyViolation checks if the error is a foreign key violation
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == ForeignKeyViolation
	}
	return false
}
