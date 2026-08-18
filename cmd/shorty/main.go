package main

import (
	"context"
	"log"
	"shorty/internal/delivery/http"
	"shorty/internal/domain"
	"shorty/internal/repository"
	"shorty/internal/repository/redis"
	"shorty/internal/service"
	"shorty/internal/service/hash"
)

func prepareRedisCache(ctx context.Context, cfg *repository.Credentials) domain.URLRepository {
	cache, err := redis.NewRedisClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to make new redis client: %v", err)
	}
	return cache
}

func prepareService(cache domain.URLRepository, hasher domain.Hasher) domain.URLService {
	tools := &domain.Tools{
		Cache:  cache,
		Hasher: hasher,
	}
	return service.New(tools)
}

func main() {
	ctx := context.Background()
	cfg := &repository.Credentials{Address: "localhost:6379", Password: "123456", DB: 0}

	generator := hash.NewMurmur()
	encoder := hash.NewBase62Hash()
	hasher := hash.NewURLHasher(generator, encoder)

	cache := prepareRedisCache(ctx, cfg)

	serv := prepareService(cache, hasher)
	go serv.Run(ctx)

	server := http.NewServer(serv)
	server.Listen()

	serv.Stop()
}
