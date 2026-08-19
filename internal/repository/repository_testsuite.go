package repository

import (
	"context"
	"shorty/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func RunTest_SaveAndGet(t *testing.T, ctx context.Context, client domain.URLRepository) {
	hash := "1"

	url := domain.URL{
		LongURL:   "google.com",
		Hash:      hash,
		Redirects: 10,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err := client.Save(ctx, &url)
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

func RunTestWait_SaveAndGet(t *testing.T, ctx context.Context, client domain.URLRepository) {
	hash := "1"

	url := domain.URL{
		LongURL:   "google.com",
		Hash:      hash,
		Redirects: 10,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(4 * time.Second),
	}

	err := client.Save(ctx, &url)
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
