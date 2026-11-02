# kvdock-go

High-performance, generic key-value caching layer for Go with zero external dependencies.

## Features

- **Generic API**: Type-safe `Get[T]`, `Set[T]`, and `GetOrSet[T]` without type assertions.
- **`MemoryAdapter`**: In-memory storage with LRU eviction and atomic TTL expiry.
- **Namespace Isolation**: Key prefixing support with no delimiter leakage.
- **Thread Safety**: Concurrent read/write with atomic counters and mutex locks.
- **Zero Runtime Dependencies**: Standard library only.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zisam-glitch/kvdock-go"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	ctx := context.Background()
	client, _ := kvdock.New(kvdock.Config{
		Prefix:     "app",
		DefaultTTL: 5 * time.Minute,
	})

	// cache-aside pattern
	user, _ := kvdock.GetOrSet(ctx, client, "user:101", func() (User, error) {
		return User{ID: 101, Name: "Alice"}, nil
	})

	fmt.Printf("user: %+v\n", user)
}
```

## Testing

```sh
go test -v -race ./...
```

## Ecosystem & Other Languages

KVDock is also available natively across multiple languages with matching design principles and zero external dependencies:

- **TypeScript / Node.js**: [`zisam-glitch/kvdock`](https://github.com/zisam-glitch/kvdock) — Reference implementation with Tiered Caching, SQLite, S3, Redis, and Cloudflare KV adapters.
- **Rust**: [`zisam-glitch/kvdock-rs`](https://github.com/zisam-glitch/kvdock-rs) — Thread-safe caching layer with monotonic `Instant` TTL.
- **Swift**: [`zisam-glitch/kvdock-swift`](https://github.com/zisam-glitch/kvdock-swift) — Actor-isolated concurrent cache with `ContinuousClock`.
- **Kotlin**: [`zisam-glitch/kvdock-kt`](https://github.com/zisam-glitch/kvdock-kt) — Thread-safe $O(1)$ LRU key-value cache for the JVM.

## License

MIT
