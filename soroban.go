package soroban

import (
	"time"
)

type Options struct {
	Domain        string
	DirectoryType string
	Directory     ServerInfo
}

type ServerInfo struct {
	Hostname string
	Port     int
}

// Service interface
type Service interface{}

// Soroban interface
type Soroban interface {
	ID() string
	Register(name string, service Service) error
	Start(seed string) error
	Stop()
	WaitForStart()
}

// Directory interface
type Directory interface {
	// TimeToLive return duration from mode.
	TimeToLive(mode string) time.Duration

	// List return all known values for this key.
	List(key string) ([]string, error)

	// Add value in key.
	// TimeToLive must be greter or equals to 1 second.
	// Multiple values can be store with the same key.
	// TTL is the same for all values.
	Add(key, value string, TTL time.Duration) error

	// Remove value from key.
	Remove(key, value string) error
}
