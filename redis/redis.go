package redis

import (
	"context"
	"log"

	cfg "github.com/ONSdigital/dis-redirect-proxy/config"
	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

type RedisClient struct {
	cfg.RedisConfig
}

// Init returns an initialised Redis client encapsulating a connection to the redis server/cluster with the given configuration
// and an error
func (r *RedisClient) Init(ctx context.Context) (disRedisClient *disRedis.Client, err error) {
	redisClient, redisClientErr := disRedis.NewClient(ctx, &disRedis.ClientConfig{
		Address: r.Address,
	})
	if redisClientErr != nil {
		log.Fatal(ctx, "failed to create dis-redis client", redisClientErr)
	}

	return redisClient, err
}

// Checker call the healthcheck of dis-redis and returns the health state of the redis instance
func (r *RedisClient) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	redisClient, err := r.Init(ctx)
	if err != nil {
		log.Fatal(ctx, "could not instantiate dis-redis client", err)
		return err
	}

	return redisClient.Checker(context.Background(), state)
}
