package kvdock

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"
)

type Config struct {
	Adapter    StorageAdapter
	Prefix     string
	DefaultTTL time.Duration
}

type Client struct {
	adapter    StorageAdapter
	prefix     string
	defaultTTL time.Duration

	hits    atomic.Int64
	misses  atomic.Int64
	sets    atomic.Int64
	deletes atomic.Int64
}

func New(cfg Config) (*Client, error) {
	if cfg.Adapter == nil {
		cfg.Adapter = NewMemoryAdapter(10000)
	}
	if cfg.DefaultTTL < 0 {
		return nil, ErrInvalidTTL
	}
	return &Client{
		adapter:    cfg.Adapter,
		prefix:     cfg.Prefix,
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

func (c *Client) resolveKey(key string) (string, error) {
	if key == "" {
		return "", ErrEmptyKey
	}
	if c.prefix == "" {
		return key, nil
	}
	return c.prefix + ":" + key, nil
}

func (c *Client) Stats() CacheStats {
	return CacheStats{
		Hits:    c.hits.Load(),
		Misses:  c.misses.Load(),
		Sets:    c.sets.Load(),
		Deletes: c.deletes.Load(),
	}
}

func Get[T any](ctx context.Context, c *Client, key string) (T, bool, error) {
	var zero T
	resolved, err := c.resolveKey(key)
	if err != nil {
		return zero, false, err
	}

	raw, found, err := c.adapter.Get(ctx, resolved)
	if err != nil {
		return zero, false, err
	}
	if !found {
		c.misses.Add(1)
		return zero, false, nil
	}

	var target T
	if err := json.Unmarshal(raw, &target); err != nil {
		return zero, false, ErrCorruptPayload
	}

	c.hits.Add(1)
	return target, true, nil
}

func Set[T any](ctx context.Context, c *Client, key string, value T, opt ...SetOptions) error {
	resolved, err := c.resolveKey(key)
	if err != nil {
		return err
	}

	options := SetOptions{TTL: c.defaultTTL}
	if len(opt) > 0 {
		options = opt[0]
	}
	if options.TTL < 0 {
		return ErrInvalidTTL
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := c.adapter.Set(ctx, resolved, bytes, options); err != nil {
		return err
	}

	c.sets.Add(1)
	return nil
}

func GetOrSet[T any](ctx context.Context, c *Client, key string, fetcher func() (T, error), opt ...SetOptions) (T, error) {
	val, found, err := Get[T](ctx, c, key)
	if err != nil {
		return val, err
	}
	if found {
		return val, nil
	}

	fresh, err := fetcher()
	if err != nil {
		return fresh, err
	}

	if err := Set[T](ctx, c, key, fresh, opt...); err != nil {
		return fresh, err
	}

	return fresh, nil
}

func (c *Client) Delete(ctx context.Context, key string) (bool, error) {
	resolved, err := c.resolveKey(key)
	if err != nil {
		return false, err
	}

	deleted, err := c.adapter.Delete(ctx, resolved)
	if err != nil {
		return false, err
	}
	if deleted {
		c.deletes.Add(1)
	}
	return deleted, nil
}

func (c *Client) Has(ctx context.Context, key string) (bool, error) {
	resolved, err := c.resolveKey(key)
	if err != nil {
		return false, err
	}
	return c.adapter.Has(ctx, resolved)
}

func (c *Client) Clear(ctx context.Context) error {
	return c.adapter.Clear(ctx)
}

func (c *Client) Close() error {
	return c.adapter.Close()
}
