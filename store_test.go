package kvdock_test

import (
	"context"
	"testing"
	"time"

	"github.com/zisam-glitch/kvdock-go"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestBasicSetAndGet(t *testing.T) {
	ctx := context.Background()
	client, err := kvdock.New(kvdock.Config{Prefix: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user := User{ID: 1, Name: "Alice"}
	if err := kvdock.Set(ctx, client, "user:1", user); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	got, found, err := kvdock.Get[User](ctx, client, "user:1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !found {
		t.Fatal("expected key to be found")
	}
	if got.ID != user.ID || got.Name != user.Name {
		t.Fatalf("expected %+v, got %+v", user, got)
	}

	stats := client.Stats()
	if stats.Sets != 1 || stats.Hits != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestCacheMiss(t *testing.T) {
	ctx := context.Background()
	client, err := kvdock.New(kvdock.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found, err := kvdock.Get[string](ctx, client, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("expected found to be false")
	}

	stats := client.Stats()
	if stats.Misses != 1 {
		t.Fatalf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestGetOrSet(t *testing.T) {
	ctx := context.Background()
	client, err := kvdock.New(kvdock.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := 0
	fetcher := func() (int, error) {
		calls++
		return 42, nil
	}

	v1, err := kvdock.GetOrSet(ctx, client, "calc", fetcher)
	if err != nil || v1 != 42 || calls != 1 {
		t.Fatalf("v1=%d calls=%d err=%v", v1, calls, err)
	}

	v2, err := kvdock.GetOrSet(ctx, client, "calc", fetcher)
	if err != nil || v2 != 42 || calls != 1 {
		t.Fatalf("v2=%d calls=%d err=%v", v2, calls, err)
	}
}

func TestTTLExpiration(t *testing.T) {
	ctx := context.Background()
	client, err := kvdock.New(kvdock.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := kvdock.Set(ctx, client, "temp", "val", kvdock.SetOptions{TTL: 25 * time.Millisecond}); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	val, found, err := kvdock.Get[string](ctx, client, "temp")
	if err != nil || !found || val != "val" {
		t.Fatalf("immediate get failed: found=%v val=%s", found, val)
	}

	time.Sleep(50 * time.Millisecond)

	_, found, err = kvdock.Get[string](ctx, client, "temp")
	if err != nil || found {
		t.Fatalf("expected expired key: found=%v", found)
	}
}

func TestLRUEviction(t *testing.T) {
	ctx := context.Background()
	adapter := kvdock.NewMemoryAdapter(2)
	client, err := kvdock.New(kvdock.Config{Adapter: adapter})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	kvdock.Set(ctx, client, "k1", 1)
	kvdock.Set(ctx, client, "k2", 2)
	// access k1 so k2 becomes LRU
	kvdock.Get[int](ctx, client, "k1")
	kvdock.Set(ctx, client, "k3", 3)

	_, found1, _ := kvdock.Get[int](ctx, client, "k1")
	_, found2, _ := kvdock.Get[int](ctx, client, "k2")
	_, found3, _ := kvdock.Get[int](ctx, client, "k3")

	if !found1 || found2 || !found3 {
		t.Fatalf("expected k2 evicted: found1=%v found2=%v found3=%v", found1, found2, found3)
	}
}

func TestValidationErrors(t *testing.T) {
	ctx := context.Background()
	client, _ := kvdock.New(kvdock.Config{})

	if err := kvdock.Set(ctx, client, "", "val"); err != kvdock.ErrEmptyKey {
		t.Fatalf("expected ErrEmptyKey, got %v", err)
	}

	if err := kvdock.Set(ctx, client, "k", "val", kvdock.SetOptions{TTL: -time.Second}); err != kvdock.ErrInvalidTTL {
		t.Fatalf("expected ErrInvalidTTL, got %v", err)
	}
}
