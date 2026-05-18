package tilehistory

import (
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

func TestHistoryRejectsNonFiniteAndZeroRetention(t *testing.T) {
	h := New[float64](0)
	if h.Add(Sample[float64]{Timestamp: time.Unix(0, 0), Value: 1}) {
		t.Fatal("expected zero-retention history to reject samples")
	}
	if got := h.DroppedSamples(); got != 1 {
		t.Fatalf("dropped samples = %d, want 1", got)
	}
}
