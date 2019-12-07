package soroban

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type Redis struct {
	rdb *redis.Client
	p   redis.Pipeliner
}

func NewRedis(options OptionRedis) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", options.Hostname, options.Port),
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	p := rdb.Pipeline()
	if p == nil {
		return nil
	}
	status := p.Ping()
	if status == nil {
		return nil
	}

	return &Redis{
		rdb: rdb,
		p:   p,
	}
}
