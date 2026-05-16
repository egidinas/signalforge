package pollqueue

import (
	"sync"
	"time"
)

const (
	// StatusNotSampled marks values that are known to the queue but have not
	// returned from the poll loop yet.
	StatusNotSampled = "not_sampled"

	// MaxManualBurst is the number of front-lane emissions allowed before a
	// normal item must be considered when normal work is queued.
	MaxManualBurst = 1
)

// KeyFunc returns the stable identity used for deduplication and latest-value
// lookups.
type KeyFunc[T any] func(T) string

// Result is the freshest observed value for one queued item.
type Result[T any, V any] struct {
	Item       T
	Value      V
	Status     string
	Error      string
	ObservedAt time.Time
}

// Queue rotates normal items in bulk-friendly chunks. EnqueueFront inserts a
// one-shot manual request ahead of normal rotation while preserving normal
// scheduler fairness.
type Queue[T any, V any] struct {
	mu        sync.Mutex
	key       KeyFunc[T]
	normal    []T
	front     []T
	normalSet map[string]struct{}
	frontSet  map[string]struct{}
	latest    map[string]Result[T, V]
	manualRun int
}

// New creates a queue and seeds Latest for all initial items.
func New[T any, V any](items []T, key KeyFunc[T]) *Queue[T, V] {
	q := &Queue[T, V]{
		key:       key,
		normalSet: map[string]struct{}{},
		frontSet:  map[string]struct{}{},
		latest:    map[string]Result[T, V]{},
	}
	for _, item := range items {
		q.Enqueue(item)
	}
	return q
}

// Enqueue adds a normal round-robin item if it is not already queued.
func (q *Queue[T, V]) Enqueue(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	key := q.itemKey(item)
	q.seedLatestLocked(item, key)
	if _, ok := q.normalSet[key]; ok {
		return
	}
	q.normalSet[key] = struct{}{}
	q.normal = append(q.normal, item)
}

// EnqueueFront adds a one-shot manual item to the front lane. Repeated manual
// requests for the same key collapse to one pending front-lane item.
func (q *Queue[T, V]) EnqueueFront(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	key := q.itemKey(item)
	q.seedLatestLocked(item, key)
	if _, ok := q.frontSet[key]; ok {
		return
	}
	q.frontSet[key] = struct{}{}
	q.front = append(q.front, item)
}

// NextChunk returns up to max items, giving front-lane requests priority while
// bounding consecutive manual emissions. The same key is emitted at most once
// per chunk.
func (q *Queue[T, V]) NextChunk(max int) []T {
	if max <= 0 {
		return nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]T, 0, max)
	emitted := map[string]struct{}{}
	normalLimit := len(q.normal)
	normalTaken := 0

	for len(out) < max && (len(q.front) > 0 || (normalTaken < normalLimit && len(q.normal) > 0)) {
		if q.shouldTakeFrontLocked(normalTaken, normalLimit) {
			if q.emitFrontLocked(&out, emitted) {
				q.manualRun++
			}
			continue
		}

		if normalTaken < normalLimit && len(q.normal) > 0 {
			normalTaken++
			if q.emitNormalLocked(&out, emitted) {
				q.manualRun = 0
			}
		}
	}
	return out
}

// Record stores a fresh result. A zero ObservedAt is replaced with time.Now.
func (q *Queue[T, V]) Record(result Result[T, V]) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if result.ObservedAt.IsZero() {
		result.ObservedAt = time.Now()
	}
	q.latest[q.itemKey(result.Item)] = result
}

// Latest returns the freshest known result for item, including not-sampled
// placeholders seeded by Enqueue and EnqueueFront.
func (q *Queue[T, V]) Latest(item T) (Result[T, V], bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	result, ok := q.latest[q.itemKey(item)]
	return result, ok
}

func (q *Queue[T, V]) seedLatestLocked(item T, key string) {
	if _, ok := q.latest[key]; ok {
		return
	}
	q.latest[key] = Result[T, V]{
		Item:   item,
		Status: StatusNotSampled,
	}
}

func (q *Queue[T, V]) itemKey(item T) string {
	if q.key == nil {
		panic("pollqueue: nil key function")
	}
	return q.key(item)
}

func (q *Queue[T, V]) shouldTakeFrontLocked(normalTaken, normalLimit int) bool {
	if len(q.front) == 0 {
		return false
	}
	if len(q.normal) == 0 || normalTaken >= normalLimit {
		return true
	}
	return q.manualRun < MaxManualBurst
}

func (q *Queue[T, V]) emitFrontLocked(out *[]T, emitted map[string]struct{}) bool {
	item := q.front[0]
	q.front = q.front[1:]
	key := q.itemKey(item)
	delete(q.frontSet, key)
	if _, ok := emitted[key]; ok {
		return false
	}
	emitted[key] = struct{}{}
	*out = append(*out, item)
	return true
}

func (q *Queue[T, V]) emitNormalLocked(out *[]T, emitted map[string]struct{}) bool {
	item := q.normal[0]
	q.normal = append(q.normal[1:], item)
	key := q.itemKey(item)
	if _, ok := emitted[key]; ok {
		return false
	}
	emitted[key] = struct{}{}
	*out = append(*out, item)
	return true
}
