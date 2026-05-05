package util

import (
	"database/sql"
	"time"
)

func NullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
