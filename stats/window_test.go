package stats

import (
	"math"
	"testing"
)

func TestWindowSummarizesValues(t *testing.T) {
	var w Window
	for _, value := range []float64{10, 12, 14} {
		if !w.Add(value) {
			t.Fatalf("value %v was rejected", value)
		}
	}

	got, ok := w.Snapshot()
	if !ok {
		t.Fatal("expected summary")
	}
	if got.Count != 3 || got.Mean != 12 || got.Min != 10 || got.Max != 14 {
		t.Fatalf("summary = %#v", got)
	}
	if math.Abs(got.StdDev-2) > 0.000001 {
		t.Fatalf("StdDev = %v, want 2", got.StdDev)
	}
}

func TestWindowIgnoresNaN(t *testing.T) {
	var w Window
	if w.Add(math.NaN()) {
		t.Fatal("NaN was accepted")
	}
	w.Add(5)

	got, ok := w.Snapshot()
	if !ok {
		t.Fatal("expected summary")
	}
	if got.Count != 1 || got.Mean != 5 || got.StdDev != 0 {
		t.Fatalf("summary = %#v", got)
	}
}

func TestWindowIgnoresInfinities(t *testing.T) {
	var w Window
	if w.Add(math.Inf(1)) {
		t.Fatal("+Inf was accepted")
	}
	if w.Add(math.Inf(-1)) {
		t.Fatal("-Inf was accepted")
	}
	if !w.Add(5) || !w.Add(7) {
		t.Fatal("finite value was rejected")
	}

	got, ok := w.Snapshot()
	if !ok {
		t.Fatal("expected summary")
	}
	if got.Count != 2 || got.Mean != 6 || got.Min != 5 || got.Max != 7 {
		t.Fatalf("summary = %#v", got)
	}
	if math.IsNaN(got.Mean) || math.IsNaN(got.StdDev) {
		t.Fatalf("summary contains NaN: %#v", got)
	}
}

func TestWindowEmptyAndReset(t *testing.T) {
	var w Window
	if _, ok := w.Snapshot(); ok {
		t.Fatal("empty window returned a summary")
	}
	w.Add(1)
	w.Reset()
	if w.Count() != 0 {
		t.Fatalf("Count after reset = %d, want 0", w.Count())
	}
	if _, ok := w.Snapshot(); ok {
		t.Fatal("reset window returned a summary")
	}
}
