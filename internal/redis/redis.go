package redis

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"

	"github.com/go-redis/redis"
)

var (
	InvalidArgsErr = errors.New("Invalid Args Error")
	ListErr        = errors.New("List Error")
	AddErr         = errors.New("Add Error")
	RemoveErr      = errors.New("Remove Error")
)

type Redis struct {
	domain string
	rdb    *redis.Client
}

func New(options soroban.ServerInfo) *Redis {
	return NewWithDomain("samourai", options)
}

func NewWithDomain(domain string, options soroban.ServerInfo) *Redis {
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
		domain: domain,
		rdb:    rdb,
	}
}

// TimeToLive return duration from mode.
func (r *Redis) TimeToLive(mode string) time.Duration {
	if len(mode) == 0 {
		mode = "default"
	}

	switch mode {
	case "short":
		return time.Minute

	case "long":
		return 5 * time.Minute

	case "normal":
		fallthrough
	case "default":
		fallthrough
	default:
		return 3 * time.Minute
	}
}

// List return all known values for this key.
func (r *Redis) List(key string) ([]string, error) {
	if len(key) == 0 {
		return nil, InvalidArgsErr
	}
	key = keyHash(r.domain, key)

	values, err := r.rdb.SMembers(key).Result()
	if err != nil {
		return nil, ListErr
	}
	if len(values) == 0 {
		return nil, nil
	}

	// sort with counter prefix
	sort.Slice(values, func(i, j int) bool {
		n1, _ := parseValue(values[i])
		n2, _ := parseValue(values[j])
		return n1 < n2
	})

	// remove counter prefix from value
	for i := 0; i < len(values); i++ {
		_, value := parseValue(values[i])
		values[i] = value
	}

	return values[:], nil
}

// Add store key with value.
// TimeToLive must be greter or equals to 1 second.
// Multiple values can be store with the same key.
// TTL is the same for all values.
func (r *Redis) Add(key, value string, TTL time.Duration) error {
	if len(key) == 0 || len(value) == 0 || TTL < time.Second {
		return InvalidArgsErr
	}

	key = keyHash(r.domain, key)
	keyCounter := countHash(r.domain, key)
	valueHashKey := valueHash(r.domain, value)

	// check if valueHashKey exists
	exists, err := r.rdb.Exists(valueHashKey).Result()
	if err != nil {
		return AddErr
	}

	// if value not exists
	if exists == 0 {
		// get next absolut counter
		counter, err := r.rdb.Incr(keyCounter).Result()
		if err != nil {
			return AddErr
		}
		// set counter in valueHashKey
		ok, err := r.rdb.Set(valueHashKey, counter, TTL).Result()
		if err != nil || ok != "OK" {
			return AddErr
		}

		// store formated value in key
		n, err := r.rdb.SAdd(key, formatValue(counter, value)).Result()
		if err != nil || n != 1 {
			return AddErr
		}
	}

	// Set or extend keys lifetime
	r.rdb.Expire(key, TTL)
	r.rdb.Expire(keyCounter, TTL)
	r.rdb.Expire(valueHashKey, TTL)

	return nil
}

func (r *Redis) Remove(key, value string) error {
	if len(key) == 0 || len(value) == 0 {
		return InvalidArgsErr
	}
	key = keyHash(r.domain, key)
	valueHashKey := valueHash(r.domain, value)

	// check if valueHashKey exists
	exists, err := r.rdb.Exists(valueHashKey).Result()
	if err != nil {
		return RemoveErr
	}
	// no error if not exists
	if exists == 0 {
		return nil
	}

	// retrieve absolut counter for formated value
	counterStr, err := r.rdb.Get(valueHashKey).Result()
	if err != nil {
		return err
	}
	counter, err := strconv.ParseInt(counterStr, 10, 64)
	if err != nil {
		return RemoveErr
	}

	// delete absolut counter
	n, err := r.rdb.Del(valueHashKey).Result()
	if err != nil || n != 1 {
		return RemoveErr
	}

	// remove formated value from key
	n, err = r.rdb.SRem(key, formatValue(counter, value)).Result()
	if err != nil || n != 1 {
		return RemoveErr
	}

	// reduce expire counter key on last remove
	n, err = r.rdb.SCard(key).Result()
	if err != nil {
		return RemoveErr
	}
	if n == 0 {
		keyCounter := countHash(r.domain, key)
		_, err = r.rdb.Expire(keyCounter, 15*time.Second).Result()
		if err != nil {
			return RemoveErr
		}
	}

	return nil
}
