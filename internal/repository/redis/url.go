package redis

import (
	"shorty/internal/domain"
	"time"
)

type redisURL struct {
	ID        int64  `redis:"id"`
	LongURL   string `redis:"long_url"`
	Redirects int64  `redis:"redirects"`
	CreatedAt int64  `redis:"created_at"`
	ExpiresAt int64  `redis:"expires_at"`
}

func (rURL *redisURL) FromDomain(url *domain.URL) {
	*rURL = redisURL{
		ID:        url.ID,
		LongURL:   url.LongURL,
		Redirects: url.Redirects,
		CreatedAt: url.CreatedAt.Unix(),
		ExpiresAt: url.ExpiresAt.Unix(),
	}
}

func (rURL *redisURL) ToDomain(hash string) *domain.URL {
	return &domain.URL{
		ID:        rURL.ID,
		LongURL:   rURL.LongURL,
		Hash:      hash,
		Redirects: rURL.Redirects,
		CreatedAt: time.Unix(rURL.CreatedAt, 0).UTC(),
		ExpiresAt: time.Unix(rURL.ExpiresAt, 0).UTC(),
	}
}
