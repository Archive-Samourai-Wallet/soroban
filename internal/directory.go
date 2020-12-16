package internal

import (
	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal/memory"
	"code.samourai.io/wallet/samourai-soroban/internal/redis"
)

type DirectoryType string

const (
	DirectoryTypeMemory DirectoryType = "directory-memory"
	DirectoryTypeRedis  DirectoryType = "directory-redis"
)

func DefaultDirectory(domain string, options soroban.ServerInfo) soroban.Directory {
	return NewDirectory(domain, DirectoryTypeMemory, options)
}

func NewDirectory(domain string, DirectoryType DirectoryType, options soroban.ServerInfo) soroban.Directory {
	switch DirectoryType {
	case DirectoryTypeMemory:
		return memory.NewWithDomain(domain, memory.DefaultCacheCapacity, memory.DefaultCacheTTL)
	case DirectoryTypeRedis:
		return redis.NewWithDomain(domain, options)
	default:
		return memory.NewWithDomain(domain, memory.DefaultCacheCapacity, memory.DefaultCacheTTL)
	}
}
