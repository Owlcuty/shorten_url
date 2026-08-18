package domain

import "time"

type URL struct {
	ID        int64
	LongURL   string
	Hash      string
	Redirects int64
	CreatedAt time.Time
	ExpiresAt time.Time
}
