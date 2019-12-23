package internal

import (
	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal/redis"
)

type DirectoryType string

const (
	DirectoryTypeRedis DirectoryType = "directory-redis"
)

func DefaultDirectory(options soroban.ServerInfo) soroban.Directory {
	return NewDirectory("samourai", DirectoryTypeRedis, options)
}

func NewDirectory(domain string, DirectoryType DirectoryType, options soroban.ServerInfo) soroban.Directory {
	switch DirectoryType {
	case DirectoryTypeRedis:
		return redis.NewWithDomain(domain, options)
	default:
		return nil
	}
}
