package kvdock

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidTTL      = errors.New("kvdock: invalid ttl, must be positive duration")
	ErrEmptyKey        = errors.New("kvdock: key cannot be empty")
	ErrKeyNotFound     = errors.New("kvdock: key not found")
	ErrCorruptPayload  = errors.New("kvdock: corrupt payload")
)

type StorageCapabilities struct {
	Batch     bool
	Streaming bool
	TTL       bool
}

type SetOptions struct {
	TTL time.Duration
}

type CacheStats struct {
	Hits    int64
	Misses  int64
	Sets    int64
	Deletes int64
}

type StorageAdapter interface {
	Capabilities() StorageCapabilities
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, opt SetOptions) error
	Delete(ctx context.Context, key string) (bool, error)
	Has(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Close() error
}
