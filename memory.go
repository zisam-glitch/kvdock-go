package kvdock

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type memoryEntry struct {
	key       string
	value     []byte
	expiresAt time.Time
}

func (e *memoryEntry) isExpired(now time.Time) bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return now.After(e.expiresAt)
}

type MemoryAdapter struct {
	mu         sync.RWMutex
	items      map[string]*list.Element
	evictList  *list.List
	maxEntries int
}

func NewMemoryAdapter(maxEntries int) *MemoryAdapter {
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	return &MemoryAdapter{
		items:      make(map[string]*list.Element),
		evictList:  list.New(),
		maxEntries: maxEntries,
	}
}

func (m *MemoryAdapter) Capabilities() StorageCapabilities {
	return StorageCapabilities{
		Batch:     false,
		Streaming: false,
		TTL:       true,
	}
}

func (m *MemoryAdapter) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	elem, exists := m.items[key]
	if !exists {
		return nil, false, nil
	}

	entry := elem.Value.(*memoryEntry)
	if entry.isExpired(time.Now()) {
		m.evictList.Remove(elem)
		delete(m.items, key)
		return nil, false, nil
	}

	m.evictList.MoveToFront(elem)
	out := make([]byte, len(entry.value))
	copy(out, entry.value)
	return out, true, nil
}

func (m *MemoryAdapter) Set(_ context.Context, key string, value []byte, opt SetOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiresAt time.Time
	if opt.TTL > 0 {
		expiresAt = time.Now().Add(opt.TTL)
	}

	valCopy := make([]byte, len(value))
	copy(valCopy, value)

	if elem, exists := m.items[key]; exists {
		m.evictList.MoveToFront(elem)
		entry := elem.Value.(*memoryEntry)
		entry.value = valCopy
		entry.expiresAt = expiresAt
		return nil
	}

	if m.evictList.Len() >= m.maxEntries {
		oldest := m.evictList.Back()
		if oldest != nil {
			m.evictList.Remove(oldest)
			oldEntry := oldest.Value.(*memoryEntry)
			delete(m.items, oldEntry.key)
		}
	}

	entry := &memoryEntry{
		key:       key,
		value:     valCopy,
		expiresAt: expiresAt,
	}
	elem := m.evictList.PushFront(entry)
	m.items[key] = elem
	return nil
}

func (m *MemoryAdapter) Delete(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	elem, exists := m.items[key]
	if !exists {
		return false, nil
	}

	m.evictList.Remove(elem)
	delete(m.items, key)
	return true, nil
}

func (m *MemoryAdapter) Has(ctx context.Context, key string) (bool, error) {
	_, found, err := m.Get(ctx, key)
	return found, err
}

func (m *MemoryAdapter) Clear(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]*list.Element)
	m.evictList.Init()
	return nil
}

func (m *MemoryAdapter) Close() error {
	return m.Clear(context.Background())
}
