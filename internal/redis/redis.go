package redis

import (
	"errors"
	"fmt"
	"sort"
	"sync"
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
	sync.Mutex
	nonce uint64

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

	// sort with nonce prefix
	sort.Slice(values, func(i, j int) bool {
		n1, _ := parseValue(values[i])
		n2, _ := parseValue(values[j])
		return n1 < n2
	})

	// remove nonce prefix frpm value
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

	nonce := r.nextNonce(value)
	_, err := r.rdb.Set(nonceHash(r.domain, value), nonce, TTL).Result()
	if err != nil {
		return AddErr
	}

	value = formatValue(value, nonce)
	_, err = r.rdb.SAdd(key, value).Result()
	if err != nil {
		return AddErr
	}

	// Set or extend key's lifetime
	r.rdb.Expire(key, TTL)

	return nil
}

func (r *Redis) Remove(key, value string) error {
	if len(key) == 0 || len(value) == 0 {
		return InvalidArgsErr
	}
	key = keyHash(r.domain, key)

	nKey := nonceHash(r.domain, value)
	nonce := r.rdb.Get(nKey).Val()
	_, err := r.rdb.Del(nKey).Result()
	if err != nil {
		return RemoveErr
	}

	_, err = r.rdb.SRem(key, formatValue(value, nonce)).Result()
	if err != nil {
		return RemoveErr
	}

	return nil
}

func (r *Redis) nextNonce(value string) string {
	r.Lock()
	defer r.Unlock()
	r.nonce++

	return fmt.Sprintf("%d", r.nonce)
}
