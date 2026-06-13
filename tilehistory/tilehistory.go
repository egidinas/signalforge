package tilehistory

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Number is the supported sample value domain.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Sample is one timestamped numeric reading.
type Sample[T Number] struct {
	Timestamp time.Time
	Value     T
}

// BucketSummary is the aggregate for one fixed interval.
type BucketSummary struct {
	IntervalStart time.Time `json:"interval_start"`
	IntervalEnd   time.Time `json:"interval_end"`
	Count         int       `json:"count"`
	Mean          float64   `json:"mean"`
	Min           float64   `json:"min"`
	Max           float64   `json:"max"`
}

// Snapshot is a tile-friendly immutable view of retained history.
type Snapshot struct {
	Anchor         time.Time       `json:"anchor"`
	Interval       time.Duration   `json:"interval"`
	Retention      time.Duration   `json:"retention"`
	Earliest       time.Time       `json:"earliest"`
	Latest         time.Time       `json:"latest"`
	DroppedSamples int             `json:"dropped_samples"`
	Count          int             `json:"count"`
	Buckets        []BucketSummary `json:"buckets"`
}

type bucket struct {
	intervalStart time.Time
	count         int
	sum           float64
	min           float64
	max           float64
}

// History retains numeric samples in second-sized buckets.
type History[T Number] struct {
	mu             sync.RWMutex
	interval       time.Duration
	retention      time.Duration
	maxBuckets     int
	buckets        []bucket
	bucketCount    int
	head           int
	droppedSamples int
	anchor         time.Time
	latest         time.Time
}

// New creates a bounded history reducer.
//
// retention is rounded down to whole intervals. Values shorter than one
// interval disable retention.
func New[T Number](retention time.Duration) *History[T] {
	return NewWithInterval[T](time.Second, retention)
}

// NewWithInterval creates a bounded history reducer with an explicit cadence.
func NewWithInterval[T Number](interval, retention time.Duration) *History[T] {
	if interval <= 0 {
		interval = time.Second
	}
	maxBuckets := 0
	if retention > 0 {
		maxBuckets = int(retention / interval)
	}
	var buckets []bucket
	if maxBuckets > 0 {
		buckets = make([]bucket, maxBuckets)
	}
	return &History[T]{
		interval:   interval,
		retention:  retention,
		maxBuckets: maxBuckets,
		buckets:    buckets,
	}
}

// Add folds one timestamped sample into the history.
//
// Non-finite values are rejected. Samples older than the currently retained
// window are dropped rather than shifting the window backwards.
func (h *History[T]) Add(sample Sample[T]) bool {
	value := toFloat64(sample.Value)
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	ts := sample.Timestamp.UTC().Truncate(h.interval)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.maxBuckets == 0 {
		h.droppedSamples++
		return false
	}
	if !h.latest.IsZero() && ts.Before(h.windowStartLocked(h.latest)) {
		h.droppedSamples++
		return false
	}
	h.latest = maxTime(h.latest, ts)
	if h.anchor.IsZero() {
		h.anchor = ts
	}
	h.advanceLocked()
	idx, found := h.findBucket(ts)
	if found {
		h.addToBucket(idx, value)
		return true
	}
	next := bucket{
		intervalStart: ts,
		count:         1,
		sum:           value,
		min:           value,
		max:           value,
	}
	if !h.insertBucketLocked(idx, next) {
		h.droppedSamples++
		return false
	}
	h.refreshAnchorLocked()
	return true
}

// Snapshot returns the current retained history in chronological order.
func (h *History[T]) Snapshot() Snapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := Snapshot{
		Anchor:         h.anchor,
		Interval:       h.interval,
		Retention:      h.retention,
		DroppedSamples: h.droppedSamples,
		Count:          0,
	}
	if h.bucketCount == 0 {
		return out
	}
	out.Earliest = h.bucketAt(0).intervalStart
	out.Latest = h.bucketAt(h.bucketCount - 1).intervalStart
	out.Buckets = make([]BucketSummary, 0, h.bucketCount)
	for i := 0; i < h.bucketCount; i++ {
		b := h.bucketAt(i)
		out.Count += b.count
		out.Buckets = append(out.Buckets, BucketSummary{
			IntervalStart: b.intervalStart,
			IntervalEnd:   b.intervalStart.Add(h.interval),
			Count:         b.count,
			Mean:          b.sum / float64(b.count),
			Min:           b.min,
			Max:           b.max,
		})
	}
	return out
}

// Len reports the number of retained intervals.
func (h *History[T]) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.bucketCount
}

// DroppedSamples reports samples that were rejected because they were outside
// the retained window or could not be represented.
func (h *History[T]) DroppedSamples() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.droppedSamples
}

// Interval returns the history cadence.
func (h *History[T]) Interval() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.interval
}

// Retention returns the configured retention window.
func (h *History[T]) Retention() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.retention
}

func (h *History[T]) advanceLocked() {
	if h.bucketCount == 0 {
		return
	}
	minStart := h.windowStartLocked(h.latest)
	for h.bucketCount > 0 {
		b := h.bucketAt(0)
		if !b.intervalStart.Before(minStart) {
			break
		}
		h.droppedSamples += b.count
		*b = bucket{}
		h.head = (h.head + 1) % h.maxBuckets
		h.bucketCount--
	}
	if h.bucketCount == 0 {
		h.head = 0
	}
	h.refreshAnchorLocked()
}

func (h *History[T]) findBucket(ts time.Time) (int, bool) {
	if h.bucketCount == 0 {
		return 0, false
	}
	last := h.bucketAt(h.bucketCount - 1).intervalStart
	if ts.Equal(last) {
		return h.bucketCount - 1, true
	}
	if ts.After(last) {
		return h.bucketCount, false
	}
	idx := sort.Search(h.bucketCount, func(i int) bool {
		return !h.bucketAt(i).intervalStart.Before(ts)
	})
	return idx, idx < h.bucketCount && h.bucketAt(idx).intervalStart.Equal(ts)
}

func (h *History[T]) addToBucket(idx int, value float64) {
	b := h.bucketAt(idx)
	b.count++
	b.sum += value
	if value < b.min {
		b.min = value
	}
	if value > b.max {
		b.max = value
	}
}

func (h *History[T]) insertBucketLocked(idx int, next bucket) bool {
	if h.maxBuckets == 0 {
		return false
	}
	if h.bucketCount == 0 {
		h.head = 0
		h.buckets[0] = next
		h.bucketCount = 1
		return true
	}
	if idx == h.bucketCount {
		if h.bucketCount == h.maxBuckets {
			oldest := h.bucketAt(0)
			h.droppedSamples += oldest.count
			*oldest = bucket{}
			h.head = (h.head + 1) % h.maxBuckets
			h.bucketCount--
		}
		h.buckets[(h.head+h.bucketCount)%h.maxBuckets] = next
		h.bucketCount++
		return true
	}
	if h.bucketCount >= h.maxBuckets {
		return false
	}
	h.linearizeLocked()
	copy(h.buckets[idx+1:h.bucketCount+1], h.buckets[idx:h.bucketCount])
	h.buckets[idx] = next
	h.bucketCount++
	return true
}

func (h *History[T]) linearizeLocked() {
	if h.head == 0 || h.bucketCount <= 1 {
		h.head = 0
		return
	}
	ordered := make([]bucket, h.bucketCount)
	for i := 0; i < h.bucketCount; i++ {
		ordered[i] = *h.bucketAt(i)
	}
	copy(h.buckets, ordered)
	h.head = 0
}

func (h *History[T]) bucketAt(logical int) *bucket {
	return &h.buckets[(h.head+logical)%h.maxBuckets]
}

func (h *History[T]) refreshAnchorLocked() {
	if h.bucketCount == 0 {
		h.anchor = time.Time{}
		return
	}
	h.anchor = h.bucketAt(0).intervalStart
}

func (h *History[T]) windowStartLocked(latest time.Time) time.Time {
	return latest.Add(-time.Duration(h.maxBuckets-1) * h.interval)
}

func toFloat64[T Number](value T) float64 {
	return float64(value)
}

func maxTime(a, b time.Time) time.Time {
	if a.IsZero() || b.After(a) {
		return b
	}
	return a
}
