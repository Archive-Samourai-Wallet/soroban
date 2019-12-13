package soroban

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

var (
	InvalidArgsErr = errors.New("Invalid Args Error")
	PutErr         = errors.New("Put Error")
	ExistsErr      = errors.New("Exists Error")
	GetErr         = errors.New("Get Error")
	DelErr         = errors.New("Del Error")
)

type Redis struct {
	rdb *redis.Client
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

	status := rdb.Ping()
	if status == nil {
		return nil
	}

	return &Redis{
		rdb: rdb,
	}
}

// Exists return true if key exists.
func (r *Redis) Exists(key string) bool {
	values, _ := r.Get(key)
	return len(values) > 0
}

// Put store key with value.
// TimeToLive must be greter or equals to 1 second.
// Multiple values can be store with the same key.
// TTL is the same for all values.
func (r *Redis) Put(key, value string, TTL time.Duration) error {
	if len(key) == 0 || len(value) == 0 || TTL < time.Second {
		return InvalidArgsErr
	}

	pipe := r.rdb.Pipeline()

	pipe.SAdd(key, value)
	pipe.Expire(key, TTL)

	_, err := pipe.Exec()

	if err != nil {
		return err
	}

	return nil
}

// Get return all known values for this key.
func (r *Redis) Get(key string) ([]string, error) {
	if len(key) == 0 {
		return nil, InvalidArgsErr
	}

	values, err := r.rdb.SMembers(key).Result()
	if err != nil {
		return nil, GetErr
	}
	if len(values) == 0 {
		return nil, nil
	}
	return values[:], nil
}

// Del remove exiting key.
// Unknown key or expired keys do not return an error.
func (r *Redis) Del(key string) error {
	if len(key) == 0 {
		return InvalidArgsErr
	}

	_, err := r.rdb.Expire(key, 0).Result()
	if err != nil {
		return DelErr
	}

	return nil
}
