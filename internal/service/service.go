package service

import (
	"context"
	"fmt"
	"shorty/internal/domain"
	"sync"
	"time"
)

type urlService struct {
	tools          *domain.Tools
	RedirectLinkCh chan string
	wg             sync.WaitGroup
}

func New(tools *domain.Tools) *urlService {
	return &urlService{tools: tools, RedirectLinkCh: make(chan string, 10000)}
}

func (s *urlService) Run(ctx context.Context) {
	s.wg.Add(1)
	defer s.wg.Done()

	for hash := range s.RedirectLinkCh {
		_ = s.incrRedirects(ctx, hash)
	}
}

func (s *urlService) Stop() {
	close(s.RedirectLinkCh)
	s.wg.Wait()
}

func (s *urlService) Shorten(ctx context.Context, longURL string, ttlDays int) (string, error) {
	// Hash longURL
	hash, err := s.tools.Hasher.Hash(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to hash %s: %w", longURL, err)
	}

	// Check if already have
	url, err := s.GetURL(ctx, hash)
	if err != nil {
		return "", err
	}
	if url != nil {
		return url.Hash, err
	}
	now := time.Now().UTC().Truncate(time.Second)
	ttl := ttlDays
	if ttl == 0 {
		ttl = 30
	}
	// Form url and save
	url = &domain.URL{
		LongURL:   longURL,
		Hash:      hash,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Duration(ttlDays) * time.Hour),
	}
	err = s.save(ctx, url)

	return url.Hash, err
}

func (s *urlService) Resolve(ctx context.Context, hash string) (*domain.URL, error) {
	url, err := s.GetURL(ctx, hash)
	if err != nil || url == nil {
		return nil, err
	}
	select {
	case s.RedirectLinkCh <- hash:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(200 * time.Millisecond):
		return nil, fmt.Errorf("Ignore redirect count, queue is full")
	}

	return url, err
}

func (s *urlService) GetURL(ctx context.Context, hash string) (*domain.URL, error) {
	if s.tools.Cache != nil {
		url, err := s.tools.Cache.GetByHash(ctx, hash)
		if err != nil {
			return nil, err
		}
		if url != nil {
			return url, err
		}
	}
	if s.tools.DB != nil {
		return s.tools.DB.GetByHash(ctx, hash)
	}
	return nil, nil
}

func (s *urlService) GetLink(ctx context.Context, hash string) (string, error) {
	url, err := s.GetURL(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("failed to get url for %s: %w", hash, err)
	}

	return url.LongURL, err
}

func (s *urlService) save(ctx context.Context, url *domain.URL) error {
	if s.tools.DB != nil {
		err := s.tools.DB.Save(ctx, url)
		if err != nil {
			return err
		}
	}
	if s.tools.Cache != nil {
		return s.tools.Cache.Save(ctx, url)
	}
	return nil
}

func (s *urlService) incrRedirects(ctx context.Context, hash string) error {
	if s.tools.Cache != nil {
		err := s.tools.Cache.IncrementRedirects(ctx, hash)
		if err != nil {
			return err
		}
	}
	if s.tools.DB != nil {
		return s.tools.DB.IncrementRedirects(ctx, hash)
	}
	return nil
}
