package redis

import (
	"context"
	"fmt"
	"shorty/internal/configuration"
	"shorty/internal/domain"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, cfg *configuration.CacheConfig) (*RedisCache, error) {
	endpoint := cfg.Address
	if cfg.Port != "" {
		endpoint += ":" + cfg.Port
	}
	client := redis.NewClient(&redis.Options{
		Addr:     endpoint,
		Username: cfg.User,
		Password: cfg.Password,
	})

	err := client.Ping(ctx).Err()

	return &RedisCache{client: client}, err
}

func (r *RedisCache) GetByHash(ctx context.Context, hash string) (*domain.URL, error) {
	link := r.client.HGetAll(ctx, "link:"+hash)
	val, err := link.Result()

	if err != nil {
		return nil, err
	}

	if len(val) == 0 {
		return nil, nil
	}

	var rURL redisURL
	if err := link.Scan(&rURL); err != nil {
		return nil, err
	}

	url := rURL.ToDomain(hash)
	redirectsVal, err := r.client.Get(ctx, "redirects:"+hash).Result()
	if err != redis.Nil {
		redirects, err := strconv.ParseInt(redirectsVal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing redirects: %w", err)
		}
		if url.Redirects < redirects {
			url.Redirects = redirects
		}
	}

	return url, nil
}

func (r *RedisCache) Save(ctx context.Context, url *domain.URL) error {
	var rURL redisURL
	rURL.FromDomain(url)
	hash := url.Hash
	val, err := r.client.Get(ctx, "redirects:"+hash).Result()
	if err != redis.Nil {
		redirects, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse redirects for %s: %w", hash, err)
		}
		if url.Redirects < redirects {
			rURL.Redirects = redirects
		}
	} else {
		err = r.client.Set(ctx, "redirects:"+hash, rURL.Redirects, 0).Err()
		if err != nil {
			return fmt.Errorf("failed to set redirects for %s: %w", hash, err)
		}
	}

	pipe := r.client.Pipeline()

	pipe.HSet(ctx, "link:"+hash, rURL).Err()
	pipe.ExpireAt(ctx, "link:"+hash, url.ExpiresAt)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set link for %s: %w", hash, err)
	}

	return err
}

func (r *RedisCache) IncrementRedirects(ctx context.Context, hash string) error {
	return r.client.Incr(ctx, "redirects:"+hash).Err()
}

func (r *RedisCache) Close() {
	r.client.Close()
}
