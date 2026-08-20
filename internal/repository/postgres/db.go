package postgres

import (
	"context"
	"fmt"
	"log"
	"shorty/internal/configuration"
	"shorty/internal/domain"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type connection struct {
	conn        *pgxpool.Pool
	stopCleanup chan struct{}
	wg          sync.WaitGroup
}

func startWatchExpired(c *connection, period time.Duration) {
	c.wg.Go(func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.deleteExpired(context.Background())
			case <-c.stopCleanup:
				return
			}
		}
	})
}

func buildConnectionStr(cfg *configuration.DatabaseConfig) string {
	connstr := "postgres://" + cfg.User + ":" + cfg.Password + "@" + cfg.Address + ":" + cfg.Port + "/" + cfg.DatabaseName + "?sslmode=disable"
	log.Printf("Connection string: %s", connstr)
	return connstr
}

func NewConnection(ctx context.Context, cfg *configuration.DatabaseConfig, cleanupPeriod time.Duration) (*connection, error) {
	c, err := pgxpool.New(ctx, buildConnectionStr(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres %s: %w", cfg.Address, err)
	}
	if err := c.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to create connection for %s: %w", cfg.Address, err)
	}
	conn := &connection{conn: c, stopCleanup: make(chan struct{})}
	startWatchExpired(conn, cleanupPeriod)
	return conn, err
}

func (c *connection) Close() {
	close(c.stopCleanup)
	c.conn.Close()
}

func CreateBaseIfNotExist(ctx context.Context, c *connection) error {
	_, err := c.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id			BIGSERIAL PRIMARY KEY,
			long_url	TEXT NOT NULL,
			hash		TEXT UNIQUE NOT NULL,
			redirects	BIGINT DEFAULT 0,
			created_at	TIMESTAMPTZ NOT NULL,
			expires_at	TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_urls_hash ON urls (hash);
        CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls (expires_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to create base: %w", err)
	}
	return err
}

func (c *connection) Save(ctx context.Context, url *domain.URL) error {
	err := c.conn.QueryRow(ctx, `INSERT INTO urls (long_url, hash, redirects, created_at, expires_at)
								VALUES ($1, $2, $3, $4, $5)
								ON CONFLICT (hash) DO UPDATE
								SET hash = EXCLUDED.hash
								RETURNING id, redirects`,
		url.LongURL, url.Hash, url.Redirects, url.CreatedAt, url.ExpiresAt).Scan(&url.ID, &url.Redirects)

	if err != nil {
		return fmt.Errorf("failed to save {%s || %s}: %w", url.Hash, url.LongURL, err)
	}
	return nil
}

func (c *connection) GetByHash(ctx context.Context, hash string) (*domain.URL, error) {
	var url domain.URL
	err := c.conn.QueryRow(ctx, `SELECT id, long_url, hash, redirects, created_at, expires_at FROM urls
										WHERE hash = $1`, hash).Scan(
		&url.ID,
		&url.LongURL,
		&url.Hash,
		&url.Redirects,
		&url.CreatedAt,
		&url.ExpiresAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get url by %s: %w", hash, err)
	}

	return &url, nil
}
func (c *connection) IncrementRedirects(ctx context.Context, hash string) error {
	_, err := c.conn.Exec(ctx, `UPDATE urls SET redirects = redirects + 1 WHERE hash = $1`, hash)
	return err
}

func (c *connection) deleteExpired(ctx context.Context) {
	_, err := c.conn.Exec(ctx, "DELETE FROM urls WHERE expires_at < NOW()")
	if err != nil {
		log.Printf("Error: failed to delete expired urls from DB: %v", err)
	}
}
