package domain

import (
	"context"
)

type URLRepository interface {
	Save(ctx context.Context, url *URL) error
	GetByHash(ctx context.Context, hash string) (*URL, error)
	IncrementRedirects(ctx context.Context, hash string) error
	Close()
}

type URLService interface {
	Run(ctx context.Context)
	Shorten(ctx context.Context, longURL string, ttlDays int) (string, error)
	Resolve(ctx context.Context, hash string) (*URL, error)
	GetURL(ctx context.Context, hash string) (*URL, error)
	GetLink(ctx context.Context, hash string) (string, error)
	Stop()
}
