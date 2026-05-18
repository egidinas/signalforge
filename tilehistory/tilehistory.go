package tilehistory

import (
	"math"
	"sort"
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
	interval       time.Duration
	retention      time.Duration
	maxBuckets     int
	buckets        []bucket
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
		if maxBuckets == 0 {
			maxBuckets = 1
		}
	}
	return &History[T]{
		interval:   interval,
		retention:  retention,
		maxBuckets: maxBuckets,
		buckets:    make([]bucket, 0, maxBuckets),
	}
}

// Add folds one timestamped sample into the history.
//
// Non-finite values are rejected. Samples older than the currently retained
// window are dropped rather than shifting the window backwards.
func (h *History[T]) Add(sample Sample[T]) bool {
	if math.IsNaN(toFloat64(sample.Value)) || math.IsInf(toFloat64(sample.Value), 0) {
		return false
	}
	ts := sample.Timestamp.UTC().Truncate(h.interval)
	if h.maxBuckets == 0 {
		h.droppedSamples++
		return false
	}
	if !h.anchor.IsZero() && ts.Before(h.anchor.Add(-time.Duration(h.maxBuckets-1)*h.interval)) {
		h.droppedSamples++
		return false
	}
	h.latest = maxTime(h.latest, ts)
	if h.anchor.IsZero() {
		h.anchor = ts
	}
	h.advanceLocked(ts)
	idx := h.findBucket(ts)
	if idx >= 0 {
		h.addToBucket(idx, toFloat64(sample.Value))
		return true
	}
	h.buckets = append(h.buckets, bucket{
		intervalStart: ts,
		count:         1,
		sum:           toFloat64(sample.Value),
		min:           toFloat64(sample.Value),
		max:           toFloat64(sample.Value),
	})
	h.sortAndTrimLocked()
	return true
}

// Snapshot returns the current retained history in chronological order.
func (h *History[T]) Snapshot() Snapshot {
	out := Snapshot{
		Anchor:         h.anchor,
		Interval:       h.interval,
		Retention:      h.retention,
		DroppedSamples: h.droppedSamples,
		Count:          0,
	}
	if len(h.buckets) == 0 {
		return out
	}
	out.Earliest = h.buckets[0].intervalStart
	out.Latest = h.buckets[len(h.buckets)-1].intervalStart
	out.Buckets = make([]BucketSummary, 0, len(h.buckets))
	for _, b := range h.buckets {
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
func (h *History[T]) Len() int { return len(h.buckets) }

// DroppedSamples reports samples that were rejected because they were outside
// the retained window or could not be represented.
func (h *History[T]) DroppedSamples() int { return h.droppedSamples }

// Interval returns the history cadence.
func (h *History[T]) Interval() time.Duration { return h.interval }

// Retention returns the configured retention window.
func (h *History[T]) Retention() time.Duration { return h.retention }

func (h *History[T]) advanceLocked(ts time.Time) {
	if len(h.buckets) == 0 {
		return
	}
	minStart := ts.Add(-time.Duration(h.maxBuckets-1) * h.interval)
	i := 0
	for i < len(h.buckets) && h.buckets[i].intervalStart.Before(minStart) {
		h.droppedSamples += h.buckets[i].count
		i++
	}
	if i > 0 {
		h.buckets = append([]bucket(nil), h.buckets[i:]...)
	}
}

func (h *History[T]) findBucket(ts time.Time) int {
	for i := range h.buckets {
		if h.buckets[i].intervalStart.Equal(ts) {
			return i
		}
	}
	return -1
}

func (h *History[T]) addToBucket(idx int, value float64) {
	b := &h.buckets[idx]
	b.count++
	b.sum += value
	if value < b.min {
		b.min = value
	}
	if value > b.max {
		b.max = value
	}
}

func (h *History[T]) sortAndTrimLocked() {
	sort.Slice(h.buckets, func(i, j int) bool {
		return h.buckets[i].intervalStart.Before(h.buckets[j].intervalStart)
	})
	if h.maxBuckets > 0 && len(h.buckets) > h.maxBuckets {
		excess := len(h.buckets) - h.maxBuckets
		for i := 0; i < excess; i++ {
			h.droppedSamples += h.buckets[i].count
		}
		h.buckets = append([]bucket(nil), h.buckets[excess:]...)
	}
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
