package redis

import (
	"context"
	"shorty/internal/configuration"
	"shorty/internal/repository"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var db int = 0

func prepareClient(t *testing.T, ctx context.Context) *RedisCache {
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
	require.NoError(t, err)

	client, err := NewRedisClient(ctx, &configuration.CacheConfig{
		Address: endpoint,
	})
	require.NoError(t, err)

	return client
}

func TestRedisCache_SaveAndGet(t *testing.T) {
	ctx := context.Background()

	client := prepareClient(t, ctx)
	defer client.Close()

	repository.RunTest_SaveAndGet(t, ctx, client)
}

func TestRedisCacheAndWait_SaveAndGet(t *testing.T) {
	ctx := context.Background()

	client := prepareClient(t, ctx)
	defer client.Close()

	repository.RunTestWait_SaveAndGet(t, ctx, client)
}
