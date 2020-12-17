package redis

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal/common"

	"github.com/go-redis/redis"
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
	return common.TimeToLive(mode)
}

// Status returs internal informations
func (r *Redis) Status() (soroban.StatusInfo, error) {
	result, err := r.rdb.Info("all").Result()
	if err != nil {
		return soroban.StatusInfo{}, err
	}

	// parse redis bulk string
	obj := make(map[string]interface{})
	sectionName := ""
	section := obj
	scanner := bufio.NewScanner(strings.NewReader(result))
	for scanner.Scan() {
		line := scanner.Text()

		// create sub-sections
		if strings.HasPrefix(line, "#") {
			sectionName = strings.Trim(line, "# ")
			sectionName = strings.Replace(sectionName, " ", "_", -1)
			sectionName = strings.ToLower(sectionName)

			// create & add new section to obj
			section = make(map[string]interface{})
			obj[sectionName] = section
			continue
		}

		// check for key:value
		toks := strings.Split(line, ":")
		if len(toks) != 2 {
			continue
		}

		// add key/value to section
		section[toks[0]] = toks[1]
	}
	// add original data to private field
	obj["_raw"] = result

	// convert to json
	data, err := json.Marshal(obj)
	if err != nil {
		return soroban.StatusInfo{}, err
	}

	// convert back to StatusInfo
	var info soroban.StatusInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return soroban.StatusInfo{}, err
	}

	return info, nil
}

// List return all known values for this key.
func (r *Redis) List(key string) ([]string, error) {
	if len(key) == 0 {
		return nil, common.InvalidArgsErr
	}
	key = common.KeyHash(r.domain, key)

	values, err := r.rdb.SMembers(key).Result()
	if err != nil {
		return nil, common.ListErr
	}
	if len(values) == 0 {
		return nil, nil
	}

	// sort with counter prefix
	sort.Slice(values, func(i, j int) bool {
		n1, _ := common.ParseValue(values[i])
		n2, _ := common.ParseValue(values[j])
		return n1 < n2
	})

	// remove counter prefix from value
	for i := 0; i < len(values); i++ {
		_, value := common.ParseValue(values[i])
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
		return common.InvalidArgsErr
	}

	key = common.KeyHash(r.domain, key)
	keyCounter := common.CountHash(r.domain, key)
	valueHashKey := common.ValueHash(r.domain, value)

	// check if valueHashKey exists
	exists, err := r.rdb.Exists(valueHashKey).Result()
	if err != nil {
		return common.AddErr
	}

	// if value not exists
	if exists == 0 {
		// get next absolut counter
		counter, err := r.rdb.Incr(keyCounter).Result()
		if err != nil {
			return common.AddErr
		}
		// set counter in valueHashKey
		ok, err := r.rdb.Set(valueHashKey, counter, TTL).Result()
		if err != nil || ok != "OK" {
			return common.AddErr
		}

		// store formated value in key
		n, err := r.rdb.SAdd(key, common.FormatValue(counter, value)).Result()
		if err != nil || n != 1 {
			return common.AddErr
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
		return common.InvalidArgsErr
	}
	key = common.KeyHash(r.domain, key)
	valueHashKey := common.ValueHash(r.domain, value)

	// check if valueHashKey exists
	exists, err := r.rdb.Exists(valueHashKey).Result()
	if err != nil {
		return common.RemoveErr
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
		return common.RemoveErr
	}

	// delete absolut counter
	n, err := r.rdb.Del(valueHashKey).Result()
	if err != nil || n != 1 {
		return common.RemoveErr
	}

	// remove formated value from key
	n, err = r.rdb.SRem(key, common.FormatValue(counter, value)).Result()
	if err != nil || n != 1 {
		return common.RemoveErr
	}

	// reduce expire counter key on last remove
	n, err = r.rdb.SCard(key).Result()
	if err != nil {
		return common.RemoveErr
	}
	if n == 0 {
		keyCounter := common.CountHash(r.domain, key)
		_, err = r.rdb.Expire(keyCounter, 15*time.Second).Result()
		if err != nil {
			return common.RemoveErr
		}
	}

	return nil
}
