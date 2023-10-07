package internal

import (
	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal/memory"
)

type DirectoryType string

const (
	DirectoryTypeMemory DirectoryType = "directory-memory"
)

func DefaultDirectory(domain string, options soroban.ServerInfo) soroban.Directory {
	return NewDirectory(domain, DirectoryTypeMemory, options)
}

func NewDirectory(domain string, DirectoryType DirectoryType, options soroban.ServerInfo) soroban.Directory {
	switch DirectoryType {
	case DirectoryTypeMemory:
		return memory.NewWithDomain(domain, memory.DefaultCacheCapacity, memory.DefaultCacheTTL)
	default:
		return memory.NewWithDomain(domain, memory.DefaultCacheCapacity, memory.DefaultCacheTTL)
	}
}
