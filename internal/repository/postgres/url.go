package postgres

import (
	"time"
)

type url struct {
	ID        int64     `db:"id"`
	LongURL   string    `db:"long_url"`
	Hash      string    `db:"hash"`
	Redirects int64     `db:"redirects"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}
