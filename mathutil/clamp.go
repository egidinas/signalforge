// Package mathutil provides generic numeric helpers.
package mathutil

import (
	"cmp"
	"time"
)

// Clamp returns v clamped to the range [lo, hi].
func Clamp[T cmp.Ordered](v, lo, hi T) T {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ClampDuration returns v clamped to the range [0, max].
// If max <= 0, it returns v (as long as v >= 0).
func ClampDuration(v, max time.Duration) time.Duration {
	if v < 0 {
		return 0
	}
	if max > 0 && v > max {
		return max
	}
	return v
}
