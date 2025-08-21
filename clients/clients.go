package clients

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/redisclient.go -pkg mock . RedisClient

// RedisClient defines the required methods for RedisClient
type RedisClient interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	GetValue(ctx context.Context, key string) (string, error)
}
