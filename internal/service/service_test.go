package service

import (
	"context"
	"fmt"
	"math/rand"
	"shorty/internal/configuration"
	"shorty/internal/domain"
	"shorty/internal/repository/redis"
	"shorty/internal/service/hash"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type urlt struct {
	longURL string
	hash    string
}

func randomString(n int) string {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = letters[rand.Intn(len(letters))]
	}

	return string(buf)
}

func generateURLs(count int, tools *domain.Tools) []urlt {
	urls := make([]urlt, count)

	for i := 0; i < count; i++ {
		urls[i].longURL = fmt.Sprintf("http://example.com/%d/%s", i, randomString(8))
		urls[i].hash, _ = tools.Hasher.Hash(urls[i].longURL)
	}

	return urls
}

func runRedisClient(t *testing.T, ctx context.Context) (*redis.RedisCache, error) {
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

	return redis.NewRedisClient(ctx, &configuration.CacheConfig{
		Address: endpoint,
	})
}

func prepare(t *testing.T, ctx context.Context, count int) ([]urlt, *domain.Tools) {
	redisCache, err := runRedisClient(t, ctx)
	require.NoError(t, err)

	generator := hash.NewMurmur()
	encoder := hash.NewBase62Hash()

	hasher := hash.NewURLHasher(generator, encoder)

	tools := &domain.Tools{
		Cache:  redisCache,
		Hasher: hasher,
	}
	urls := generateURLs(count, tools)
	return urls, tools
}

func TestGeneral_Service(t *testing.T) {
	ctx := context.Background()
	count := 10000
	urls, tools := prepare(t, ctx, count)
	t.Log("Prepare [x]")

	serv := New(tools)
	go serv.Run(ctx)
	t.Log("Run [x]")

	for _, u := range urls {
		hash, err := serv.Shorten(ctx, u.longURL, 30)
		require.NoError(t, err)
		require.Equal(t, u.hash, hash)
	}
	t.Log("Shortened [x]")

	expectedRedirects := make(map[string]int64)

	var mu sync.Mutex

	t.Log("Groups []")

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := rand.Intn(len(urls))
			longURL := urls[id].longURL
			hash := urls[id].hash
			url, err := serv.Resolve(ctx, hash)
			require.NoError(t, err)
			mu.Lock()
			expectedRedirects[hash]++
			mu.Unlock()
			require.Equal(t, longURL, url.LongURL)
		}()
	}
	wg.Wait()
	serv.Stop()
	for _, u := range urls {
		url, err := serv.GetURL(ctx, u.hash)
		require.NoError(t, err)
		require.NotNil(t, url)
		require.Equal(t, expectedRedirects[u.hash], url.Redirects)
	}
}
