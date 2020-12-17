package memory

import (
	"sync"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal/common"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/arc"
)

const (
	DefaultCacheTTL      time.Duration = 15 * time.Minute
	DefaultCacheCapacity int           = 100000
)

type Memory struct {
	domain string
	cache  libcache.Cache
	mtx    sync.Mutex
}

func New(count int, ttl time.Duration) *Memory {
	return NewWithDomain("samourai", count, ttl)
}

func NewWithDomain(domain string, count int, ttl time.Duration) *Memory {
	cache := libcache.ARC.NewUnsafe(count)
	cache.SetTTL(ttl)

	return &Memory{
		domain: domain,
		cache:  cache,
	}
}

// Status returs internal informations
func (m *Memory) Status() (soroban.StatusInfo, error) {
	return soroban.StatusInfo{}, nil
}

// TimeToLive return duration from mode.
func (m *Memory) TimeToLive(mode string) time.Duration {
	return common.TimeToLive(mode)
}

// List return all known values for this key.
func (m *Memory) List(key string) ([]string, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if len(key) == 0 {
		return nil, common.InvalidArgsErr
	}

	key = common.KeyHash(m.domain, key)

	list := getKeyList(m.cache, key)
	result := make([]string, 0, len(list.values))
	for _, entry := range list.values {
		result = append(result, entry.value)
	}

	// keep non-expired values
	purgeKeyList(list, now())

	return result, nil
}

// Add value in key.
// TimeToLive must be greter or equals to 1 second.
// Multiple values can be store with the same key.
// TTL is the same for all values.
func (m *Memory) Add(key, value string, TTL time.Duration) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if len(key) == 0 || len(value) == 0 || TTL < time.Second {
		return common.InvalidArgsErr
	}

	key = common.KeyHash(m.domain, key)

	list := getKeyList(m.cache, key)
	list.TTL = TTL

	now := now()
	expireOn := now.Add(TTL)

	exists, pos := contains(list.values, value)
	if !exists {
		// add new value
		list.values = append(list.values, &valueEntry{
			value:    value,
			expireOn: expireOn,
		})
	} else {
		// update value expireOn
		list.values[pos].expireOn = expireOn
	}

	// keep non-expired values
	purgeKeyList(list, now)

	m.cache.StoreWithTTL(key, list, list.TTL)

	return nil
}

// Remove value from key.
func (m *Memory) Remove(key, value string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if len(key) == 0 {
		return common.InvalidArgsErr
	}

	key = common.KeyHash(m.domain, key)

	list := getKeyList(m.cache, key)
	if _, pos := contains(list.values, value); pos != -1 {
		list.values = remove(list.values, pos)
		m.cache.StoreWithTTL(key, list, list.TTL)
	}

	// keep non-expired values
	purgeKeyList(list, now())

	if len(list.values) == 0 {
		m.cache.Delete(key)
	}
	return nil
}

type valueEntry struct {
	expireOn time.Time
	value    string
}

type keyList struct {
	TTL    time.Duration
	values []*valueEntry
}

func getKeyList(cache libcache.Cache, key string) *keyList {
	if cache.Contains(key) {
		if entry, ok := cache.Load(key); ok {
			switch result := entry.(type) {
			case *keyList:
				return result
			}
		}
	}
	return &keyList{}
}

func purgeKeyList(list *keyList, limit time.Time) {
	values := list.values[:0]
	for _, value := range list.values {
		if value.expireOn.Before(limit) {
			continue
		}
		values = append(values, value)
	}
	list.values = values[:]
}

func contains(slice []*valueEntry, value string) (bool, int) {
	for i, entry := range slice {
		if entry.value == value {
			return true, i
		}
	}
	return false, -1
}

func remove(slice []*valueEntry, s int) []*valueEntry {
	return append(slice[:s], slice[s+1:]...)
}

func now() time.Time {
	return time.Now().Truncate(time.Millisecond).UTC()
}
