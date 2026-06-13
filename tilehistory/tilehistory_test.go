package tilehistory

import (
	"sync"
	"testing"
	"time"
)

func TestHistoryAggregatesBySecond(t *testing.T) {
	h := New[float64](5 * time.Second)
	base := time.Date(2026, 5, 18, 12, 0, 0, 250_000_000, time.UTC)

	if !h.Add(Sample[float64]{Timestamp: base, Value: 2}) {
		t.Fatal("first sample rejected")
	}
	if !h.Add(Sample[float64]{Timestamp: base.Add(300 * time.Millisecond), Value: 4}) {
		t.Fatal("second sample rejected")
	}
	if !h.Add(Sample[float64]{Timestamp: base.Add(1100 * time.Millisecond), Value: 10}) {
		t.Fatal("third sample rejected")
	}

	snap := h.Snapshot()
	if got, want := snap.Count, 3; got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
	if got, want := len(snap.Buckets), 2; got != want {
		t.Fatalf("bucket count = %d, want %d", got, want)
	}
	first := snap.Buckets[0]
	if got, want := first.Count, 2; got != want {
		t.Fatalf("first bucket count = %d, want %d", got, want)
	}
	if got, want := first.Mean, 3.0; got != want {
		t.Fatalf("first bucket mean = %v, want %v", got, want)
	}
	if got, want := first.Min, 2.0; got != want {
		t.Fatalf("first bucket min = %v, want %v", got, want)
	}
	if got, want := first.Max, 4.0; got != want {
		t.Fatalf("first bucket max = %v, want %v", got, want)
	}
	if got, want := first.IntervalStart, base.Truncate(time.Second); !got.Equal(want) {
		t.Fatalf("first bucket start = %s, want %s", got, want)
	}
	second := snap.Buckets[1]
	if got, want := second.Count, 1; got != want {
		t.Fatalf("second bucket count = %d, want %d", got, want)
	}
	if got, want := snap.Earliest, base.Truncate(time.Second); !got.Equal(want) {
		t.Fatalf("earliest = %s, want %s", got, want)
	}
	if got, want := snap.Latest, base.Add(1*time.Second).Truncate(time.Second); !got.Equal(want) {
		t.Fatalf("latest = %s, want %s", got, want)
	}
}

func TestHistoryBoundsRetention(t *testing.T) {
	h := New[float64](2 * time.Second)
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		if !h.Add(Sample[float64]{Timestamp: base.Add(time.Duration(i) * time.Second), Value: float64(i + 1)}) {
			t.Fatalf("sample %d rejected", i)
		}
	}
	snap := h.Snapshot()
	if got, want := len(snap.Buckets), 2; got != want {
		t.Fatalf("bucket count = %d, want %d", got, want)
	}
	if got := h.DroppedSamples(); got != 2 {
		t.Fatalf("dropped samples = %d, want 2", got)
	}
	if got, want := snap.Buckets[0].IntervalStart, base.Add(2*time.Second); !got.Equal(want) {
		t.Fatalf("oldest retained bucket = %s, want %s", got, want)
	}
}

func TestHistoryRejectsSamplesOlderThanLatestRetainedWindow(t *testing.T) {
	h := New[float64](2 * time.Second)
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		if !h.Add(Sample[float64]{Timestamp: base.Add(time.Duration(i) * time.Second), Value: float64(i + 1)}) {
			t.Fatalf("sample %d rejected", i)
		}
	}
	before := h.DroppedSamples()
	if h.Add(Sample[float64]{Timestamp: base, Value: 99}) {
		t.Fatal("expected stale sample outside retained window to be rejected")
	}
	if got, want := h.DroppedSamples(), before+1; got != want {
		t.Fatalf("dropped samples = %d, want %d", got, want)
	}
	snap := h.Snapshot()
	for _, bucket := range snap.Buckets {
		if bucket.IntervalStart.Equal(base) {
			t.Fatal("stale bucket was retained after rejection")
		}
	}
}

func TestHistoryAcceptsLateSampleInsideRetainedWindow(t *testing.T) {
	h := New[float64](3 * time.Second)
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	for _, offset := range []time.Duration{0, time.Second, 3 * time.Second} {
		if !h.Add(Sample[float64]{Timestamp: base.Add(offset), Value: 1}) {
			t.Fatalf("sample at %s rejected", offset)
		}
	}
	if !h.Add(Sample[float64]{Timestamp: base.Add(2 * time.Second), Value: 7}) {
		t.Fatal("late sample inside retained window was rejected")
	}
	snap := h.Snapshot()
	if got, want := len(snap.Buckets), 3; got != want {
		t.Fatalf("bucket count = %d, want %d", got, want)
	}
	if got, want := snap.Buckets[0].IntervalStart, base.Add(time.Second); !got.Equal(want) {
		t.Fatalf("earliest retained bucket = %s, want %s", got, want)
	}
}

func TestHistoryRejectsNonFiniteAndZeroRetention(t *testing.T) {
	h := New[float64](0)
	if h.Add(Sample[float64]{Timestamp: time.Unix(0, 0), Value: 1}) {
		t.Fatal("expected zero-retention history to reject samples")
	}
	if got := h.DroppedSamples(); got != 1 {
		t.Fatalf("dropped samples = %d, want 1", got)
	}
}

func TestHistoryShortRetentionDisablesRetention(t *testing.T) {
	h := NewWithInterval[float64](time.Second, 500*time.Millisecond)
	if h.Add(Sample[float64]{Timestamp: time.Unix(0, 0), Value: 1}) {
		t.Fatal("expected sub-interval retention to reject samples")
	}
	if got := h.DroppedSamples(); got != 1 {
		t.Fatalf("dropped samples = %d, want 1", got)
	}
	if got := h.Len(); got != 0 {
		t.Fatalf("len = %d, want 0", got)
	}
}

func TestHistoryConcurrentAddSnapshotDoesNotRace(t *testing.T) {
	h := New[float64](30 * time.Second)
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 250; i++ {
				h.Add(Sample[float64]{
					Timestamp: base.Add(time.Duration((worker*250)+i) * time.Millisecond),
					Value:     float64(i),
				})
			}
		}(worker)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			_ = h.Snapshot()
			_ = h.Len()
			_ = h.DroppedSamples()
		}
	}()
	wg.Wait()
}

func BenchmarkHistoryAddSequentialLargeRetention(b *testing.B) {
	const retained = 65536
	h := NewWithInterval[float64](time.Second, retained*time.Second)
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(Sample[float64]{Timestamp: base.Add(time.Duration(i) * time.Second), Value: float64(i)})
	}
}

func BenchmarkHistoryAddOutOfOrderWithinWindow(b *testing.B) {
	const retained = 4096
	h := NewWithInterval[float64](time.Second, retained*time.Second)
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	for i := 0; i < retained; i += 2 {
		h.Add(Sample[float64]{Timestamp: base.Add(time.Duration(i) * time.Second), Value: float64(i)})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := time.Duration((i*2+1)%retained) * time.Second
		h.Add(Sample[float64]{Timestamp: base.Add(offset), Value: float64(i)})
	}
}
