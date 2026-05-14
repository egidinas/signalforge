package stability

import (
	"math"
	"testing"
	"time"
)

func TestEvaluateLinearDriftWithinLimit(t *testing.T) {
	base := time.Unix(1000, 0)
	samples := []StabilityEvent{
		{Timestamp: base, Value: 10.0},
		{Timestamp: base.Add(30 * time.Minute), Value: 10.1},
		{Timestamp: base.Add(time.Hour), Value: 10.2},
	}
	ok, slope := EvaluateLinearDrift(samples, 0.25)
	if !ok {
		t.Fatalf("expected slope within limit, got slope=%v", slope)
	}
	if math.Abs(slope-0.2) > 1e-9 {
		t.Fatalf("unexpected slope: got %v want 0.2", slope)
	}
}

func TestEvaluateLinearDriftExceedsLimit(t *testing.T) {
	base := time.Unix(1000, 0)
	samples := []StabilityEvent{
		{Timestamp: base, Value: 10.0},
		{Timestamp: base.Add(time.Hour), Value: 11.0},
	}
	ok, slope := EvaluateLinearDrift(samples, 0.5)
	if ok {
		t.Fatalf("expected drift to exceed limit")
	}
	if math.Abs(slope-1.0) > 1e-9 {
		t.Fatalf("unexpected slope: got %v want 1.0", slope)
	}
}

func TestEvaluateLinearDriftNegativeSlope(t *testing.T) {
	base := time.Unix(1000, 0)
	samples := []StabilityEvent{
		{Timestamp: base, Value: 12.0},
		{Timestamp: base.Add(time.Hour), Value: 11.0},
	}
	ok, slope := EvaluateLinearDrift(samples, 0.5)
	if ok {
		t.Fatalf("expected negative drift magnitude to exceed limit")
	}
	if math.Abs(slope+1.0) > 1e-9 {
		t.Fatalf("unexpected slope: got %v want -1.0", slope)
	}
}

func TestEvaluateDeviationFromTarget(t *testing.T) {
	base := time.Unix(1000, 0)
	samples := []StabilityEvent{
		{Timestamp: base, Value: 19.7},
		{Timestamp: base.Add(time.Second), Value: 20.2},
	}
	ok, deviation := EvaluateDeviationFromTarget(samples, 20.0, 0.25)
	if !ok {
		t.Fatalf("expected deviation to be within limit")
	}
	if math.Abs(deviation-0.2) > 1e-9 {
		t.Fatalf("unexpected deviation: got %v want 0.2", deviation)
	}
}

func TestEvaluateNofMGroupPolicy(t *testing.T) {
	states := map[string]SignalState{
		"a": {Status: StatusStable},
		"b": {Status: StatusStable},
		"c": {Status: StatusUnstable},
	}
	ok, stable := EvaluateNofMGroupPolicy(states, 2)
	if !ok || stable != 2 {
		t.Fatalf("expected 2 stable signals to satisfy 2-of-3, got ok=%v stable=%d", ok, stable)
	}
}
