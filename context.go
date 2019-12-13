package soroban

import (
	"context"
)

const (
	SordobanRedisKey = "soroban-redis"
)

func RedisFromContext(ctx context.Context) *Redis {
	redis, _ := ctx.Value(SordobanRedisKey).(*Redis)
	return redis
}
