package redis

import (
	"context"
	"shorty/internal/domain"
	"shorty/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var db int = 0

func TestRedisCache_SaveAndGet(t *testing.T) {
	ctx := context.Background()

	redisCont, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(t, redisCont)
	require.NoError(t, err)

	endpoint, err := redisCont.Endpoint(ctx, "")
	if err != nil {
		t.Error(err)
	}

	client, err := NewRedisClient(ctx, &repository.Credentials{
		Address: endpoint,
		DB:      db,
	})
	if err != nil {
		t.Error(err)
	}

	hash := "1"

	url := domain.URL{
		LongURL:   "google.com",
		Hash:      hash,
		Redirects: 10,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err = client.Save(ctx, &url)
	require.NoError(t, err)

	gotURL, err := client.GetByHash(ctx, hash)
	require.NoError(t, err)
	require.NotNil(t, gotURL)

	require.Equal(t, url.ID, gotURL.ID)
	require.Equal(t, url.LongURL, gotURL.LongURL)
	require.Equal(t, url.Redirects, gotURL.Redirects)
	require.Equal(t, url.CreatedAt.Unix(), gotURL.CreatedAt.Unix())
	require.Equal(t, url.ExpiresAt.Unix(), gotURL.ExpiresAt.Unix())

	err = client.IncrementRedirects(ctx, hash)
	require.NoError(t, err)

	updatedURL, err := client.GetByHash(ctx, hash)
	require.NoError(t, err)
	require.NotNil(t, updatedURL)

	require.Equal(t, url.Redirects+1, updatedURL.Redirects)
}

func TestRedisCacheAndWait_SaveAndGet(t *testing.T) {
	ctx := context.Background()

	redisCont, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(t, redisCont)
	require.NoError(t, err)

	endpoint, err := redisCont.Endpoint(ctx, "")
	if err != nil {
		t.Error(err)
	}

	client, err := NewRedisClient(ctx, &repository.Credentials{
		Address: endpoint,
		DB:      db,
	})
	if err != nil {
		t.Error(err)
	}

	hash := "1"

	url := domain.URL{
		LongURL:   "google.com",
		Hash:      hash,
		Redirects: 10,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Second),
	}

	err = client.Save(ctx, &url)
	require.NoError(t, err)

	gotURL, err := client.GetByHash(ctx, hash)
	require.NoError(t, err)
	require.NotNil(t, gotURL)

	require.Equal(t, url.ID, gotURL.ID)
	require.Equal(t, url.LongURL, gotURL.LongURL)
	require.Equal(t, url.Redirects, gotURL.Redirects)
	require.Equal(t, url.CreatedAt.Unix(), gotURL.CreatedAt.Unix())
	require.Equal(t, url.ExpiresAt.Unix(), gotURL.ExpiresAt.Unix())

	err = client.IncrementRedirects(ctx, hash)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	updatedURL, err := client.GetByHash(ctx, hash)
	require.NoError(t, err)
	require.Nil(t, updatedURL)
}
