package mathutil

import (
	"testing"
	"time"
)

func TestClampInt(t *testing.T) {
	cases := []struct{ v, lo, hi, want int }{
		{5, 0, 10, 5},
		{-3, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 0, 0},
	}
	for _, tc := range cases {
		if got := Clamp(tc.v, tc.lo, tc.hi); got != tc.want {
			t.Errorf("Clamp(%d,%d,%d) = %d, want %d", tc.v, tc.lo, tc.hi, got, tc.want)
		}
	}
}

func TestClampFloat64(t *testing.T) {
	if got := Clamp(0.5, 0.0, 1.0); got != 0.5 {
		t.Errorf("got %v", got)
	}
	if got := Clamp(-1.0, 0.0, 1.0); got != 0.0 {
		t.Errorf("got %v", got)
	}
	if got := Clamp(2.0, 0.0, 1.0); got != 1.0 {
		t.Errorf("got %v", got)
	}
}

func TestClampDuration(t *testing.T) {
	cases := []struct {
		v, max time.Duration
		want   time.Duration
	}{
		{5 * time.Second, 10 * time.Second, 5 * time.Second},
		{-time.Second, 10 * time.Second, 0},
		{15 * time.Second, 10 * time.Second, 10 * time.Second},
		{5 * time.Second, 0, 5 * time.Second},
	}
	for _, tc := range cases {
		if got := ClampDuration(tc.v, tc.max); got != tc.want {
			t.Errorf("ClampDuration(%v,%v) = %v, want %v", tc.v, tc.max, got, tc.want)
		}
	}
}
