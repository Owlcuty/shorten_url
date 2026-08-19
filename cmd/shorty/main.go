package main

import (
	"context"
	"log"
	"shorty/internal/delivery/http"
	"shorty/internal/domain"
	"shorty/internal/repository"
	"shorty/internal/repository/postgres"
	"shorty/internal/repository/redis"
	"shorty/internal/service"
	"shorty/internal/service/hash"
	"time"
)

func preparePostgresDB(ctx context.Context, cfg *repository.Credentials, cleanupPeriod time.Duration) domain.URLRepository {
	log.Printf("preparePostgresDB for %s", cfg.Address)
	db, err := postgres.NewConnection(ctx, cfg, "shorty_db", cleanupPeriod)
	if err != nil {
		log.Fatalf("failed to start postgres connection: %v", err)
	}
	postgres.CreateBaseIfNotExist(ctx, db)
	return db
}

func prepareRedisCache(ctx context.Context, cfg *repository.Credentials) domain.URLRepository {
	cache, err := redis.NewRedisClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to start new redis client: %v", err)
	}
	return cache
}

func prepareService(db domain.URLRepository, cache domain.URLRepository, hasher domain.Hasher) domain.URLService {
	tools := &domain.Tools{
		DB:     db,
		Cache:  cache,
		Hasher: hasher,
	}
	return service.New(tools)
}

func main() {
	ctx := context.Background()
	redisCfg := &repository.Credentials{Address: "localhost:6379"}
	postgresCfg := &repository.Credentials{
		Address:  "localhost:5433",
		User:     "shorty_user",
		Password: "shorty_pass",
	}

	generator := hash.NewMurmur()
	encoder := hash.NewBase62Hash()
	hasher := hash.NewURLHasher(generator, encoder)

	db := preparePostgresDB(ctx, postgresCfg, time.Hour)
	if db != nil {
		defer db.Close()
	}

	cache := prepareRedisCache(ctx, redisCfg)
	if cache != nil {
		defer cache.Close()
	}

	serv := prepareService(db, cache, hasher)
	go serv.Run(ctx)

	server := http.NewServer(serv)
	server.Listen()

	serv.Stop()
}
