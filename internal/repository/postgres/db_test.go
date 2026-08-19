package postgres

import (
	"context"
	"shorty/internal/configuration"
	"shorty/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	contpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func prepareClient(t *testing.T, ctx context.Context) *connection {
	dbName := "db"
	username := "user"
	password := "pass"
	address := "localhost"
	port := "5432"
	cfg := &configuration.DatabaseConfig{
		Type:         "postgres",
		Address:      address,
		User:         username,
		Password:     password,
		Port:         port,
		DatabaseName: dbName,
	}
	postgCont, err := contpostgres.Run(
		ctx, "postgres:latest",
		contpostgres.WithDatabase(dbName),
		contpostgres.WithUsername(username),
		contpostgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(port+"/tcp").WithStartupTimeout(20*time.Second),
			wait.ForLog("database system is ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(t, postgCont)
	require.NoError(t, err)

	connStr, err := postgCont.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	connPort, err := postgCont.MappedPort(ctx, "5432")
	require.NoError(t, err)
	cfg.Port = connPort.Port()
	require.Equal(t, connStr, buildConnectionStr(cfg))

	client, err := NewConnection(ctx, cfg, 500*time.Millisecond)
	require.NoError(t, err)

	err = CreateBaseIfNotExist(ctx, client)
	require.NoError(t, err)

	return client
}

func TestPostgres_SaveAndGet(t *testing.T) {
	ctx := context.Background()

	client := prepareClient(t, ctx)
	defer client.Close()

	repository.RunTest_SaveAndGet(t, ctx, client)
}

func TestPostgresAndWait_SaveAndGet(t *testing.T) {
	ctx := context.Background()

	client := prepareClient(t, ctx)
	defer client.Close()

	repository.RunTestWait_SaveAndGet(t, ctx, client)
}
