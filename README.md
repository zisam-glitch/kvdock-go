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

## License

MIT
