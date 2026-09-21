package db

import "github.com/redis/go-redis/v9"

// NewRedisClient connects to Redis. It's used for two independent
// purposes in this service: caching entitlement reads, and de-duping
// writes via idempotency keys (see internal/handlers/subscription.go).
func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
