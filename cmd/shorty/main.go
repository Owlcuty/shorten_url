package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"shorty/internal/configuration"
	"shorty/internal/delivery"
	"shorty/internal/domain"
	"shorty/internal/repository/postgres"
	"shorty/internal/repository/redis"
	"shorty/internal/service"
	"shorty/internal/service/hash"
	"time"
)

func prepareCacheByConfig(ctx context.Context, cfg *configuration.CacheConfig) domain.URLRepository {
	switch cfg.Type {
	case "redis":
		return prepareRedisCache(ctx, cfg)
	default:
		return nil
	}
}

func prepareDBByConfig(ctx context.Context, cfg *configuration.DatabaseConfig) domain.URLRepository {
	switch cfg.Type {
	case "postgres":
		return preparePostgresDB(ctx, cfg, time.Hour)
	default:
		return nil
	}
}

func prepareHasherConfig(cfg *configuration.HasherConfig) domain.Hasher {
	var generator domain.Generator
	var encoder domain.Encoder
	switch cfg.Generator {
	case "murmur":
		generator = hash.NewMurmur()
	default:
		generator = hash.NewMurmur()
	}
	switch cfg.Encoder {
	case "base62":
		encoder = hash.NewBase62Hash()
	default:
		encoder = hash.NewBase62Hash()
	}
	return hash.NewURLHasher(generator, encoder)
}

func preparePostgresDB(ctx context.Context, cfg *configuration.DatabaseConfig, cleanupPeriod time.Duration) domain.URLRepository {
	log.Printf("preparePostgresDB for %s", cfg.Address)
	db, err := postgres.NewConnection(ctx, cfg, cleanupPeriod)
	if err != nil {
		log.Fatalf("failed to start postgres connection: %v", err)
	}
	postgres.CreateBaseIfNotExist(ctx, db)
	return db
}

func prepareRedisCache(ctx context.Context, cfg *configuration.CacheConfig) domain.URLRepository {
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

var confPath = flag.String("c", "", "Path to configuration file")

func main() {
	ctx := context.Background()

	flag.Parse()
	config, err := configuration.ParseFile(*confPath)
	if err != nil {
		log.Fatalf("failed to open configuration file, its essential (filename: %s): %v", *confPath, err)
	}

	hasher := prepareHasherConfig(&config.Hasher)

	db := prepareDBByConfig(ctx, &config.Repo.DB)
	if db != nil {
		defer db.Close()
	} else {
		log.Println("Warning: no DB in use")
	}

	cache := prepareCacheByConfig(ctx, &config.Repo.Cache)
	if cache != nil {
		defer cache.Close()
	} else {
		log.Println("Warning: no Cache in use")
	}

	if db == nil && cache == nil {
		log.Fatal("Missing both DB and Cache. Can't work without database. Check configuration or connection")
	}
	serv := prepareService(db, cache, hasher)
	go serv.Run(ctx)
	defer serv.Stop()

	server := delivery.NewServer(serv, &config.Http)
	err = server.Listen()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed listen server: %v", err)
	}
}
