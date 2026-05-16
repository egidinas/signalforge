package ringbuf

import "sync"

// Sizer estimates the retained byte cost of an item.
type Sizer[T any] func(T) int

type entry[T any] struct {
	item T
	size int
}

// Buffer retains items within a fixed byte budget, evicting the oldest items
// first. It is intended for RAM rings; durable file-backed rings should wrap it
// rather than sharing storage policy here.
type Buffer[T any] struct {
	mu       sync.Mutex
	maxBytes int
	sizer    Sizer[T]
	entries  []entry[T]
	head     int
	len      int
	bytes    int
}

// New creates a byte-budgeted buffer. Items with non-positive estimated size or
// size greater than maxBytes are dropped by Push.
func New[T any](maxBytes int, sizer Sizer[T]) *Buffer[T] {
	return &Buffer[T]{
		maxBytes: maxBytes,
		sizer:    sizer,
		entries:  make([]entry[T], 0),
	}
}

// Push appends item and evicts oldest entries until the budget fits. It returns
// false when the buffer is disabled, the item is too large, or the sizer rejects
// the item with a non-positive size.
func (b *Buffer[T]) Push(item T) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.maxBytes <= 0 || b.sizer == nil {
		return false
	}
	size := b.sizer(item)
	if size <= 0 || size > b.maxBytes {
		return false
	}
	for b.len > 0 && b.bytes+size > b.maxBytes {
		b.evictOldestLocked()
	}
	b.ensureCapacityLocked(b.len + 1)
	tail := (b.head + b.len) % cap(b.entries)
	b.entries[tail] = entry[T]{item: item, size: size}
	b.len++
	b.bytes += size
	return true
}

// Snapshot returns retained items in insertion order without clearing them.
func (b *Buffer[T]) Snapshot() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.snapshotLocked()
}

// Drain returns retained items in insertion order and clears the buffer.
func (b *Buffer[T]) Drain() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := b.snapshotLocked()
	for i := 0; i < b.len; i++ {
		idx := (b.head + i) % cap(b.entries)
		var zero entry[T]
		b.entries[idx] = zero
	}
	b.entries = b.entries[:0]
	b.head = 0
	b.len = 0
	b.bytes = 0
	return out
}

// Len returns the number of retained items.
func (b *Buffer[T]) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.len
}

// Bytes returns the estimated retained byte usage.
func (b *Buffer[T]) Bytes() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.bytes
}

// MaxBytes returns the configured byte budget.
func (b *Buffer[T]) MaxBytes() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.maxBytes
}

func (b *Buffer[T]) snapshotLocked() []T {
	if b.len == 0 {
		return []T{}
	}
	out := make([]T, 0, b.len)
	for i := 0; i < b.len; i++ {
		idx := (b.head + i) % cap(b.entries)
		out = append(out, b.entries[idx].item)
	}
	return out
}

func (b *Buffer[T]) evictOldestLocked() {
	if b.len == 0 {
		return
	}
	b.bytes -= b.entries[b.head].size
	var zero entry[T]
	b.entries[b.head] = zero
	b.head = (b.head + 1) % cap(b.entries)
	b.len--
	if b.len == 0 {
		b.head = 0
		b.entries = b.entries[:0]
	}
}

func (b *Buffer[T]) ensureCapacityLocked(minCapacity int) {
	if cap(b.entries) >= minCapacity {
		if len(b.entries) < cap(b.entries) {
			b.entries = b.entries[:cap(b.entries)]
		}
		return
	}
	newCap := cap(b.entries)
	if newCap == 0 {
		newCap = 1
	}
	for newCap < minCapacity {
		newCap *= 2
	}
	next := make([]entry[T], newCap)
	for i := 0; i < b.len; i++ {
		idx := (b.head + i) % cap(b.entries)
		next[i] = b.entries[idx]
	}
	b.entries = next
	b.head = 0
}
