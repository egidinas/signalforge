package stats

import "math"

// Summary is an immutable snapshot of a numeric window.
//
// StdDev is the sample standard deviation. It is zero for a single-value
// window.
type Summary struct {
	Count  int
	Mean   float64
	Min    float64
	Max    float64
	StdDev float64
}

// Window accumulates streaming numeric statistics with Welford's algorithm.
// NaN values are ignored so downstream reducers can pass sparse streams without
// special casing.
type Window struct {
	count int
	mean  float64
	m2    float64
	min   float64
	max   float64
}

// Add folds one value into the window. It returns false when the value was not
// accepted, currently only for non-finite values.
func (w *Window) Add(value float64) bool {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	if w.count == 0 {
		w.min = value
		w.max = value
	} else {
		if value < w.min {
			w.min = value
		}
		if value > w.max {
			w.max = value
		}
	}
	w.count++
	delta := value - w.mean
	w.mean += delta / float64(w.count)
	w.m2 += delta * (value - w.mean)
	return true
}

// Reset clears the accumulated window.
func (w *Window) Reset() {
	*w = Window{}
}

// Count returns the number of accepted values.
func (w Window) Count() int {
	return w.count
}

// Snapshot returns the current summary. The boolean is false for an empty
// window.
func (w Window) Snapshot() (Summary, bool) {
	if w.count == 0 {
		return Summary{}, false
	}
	s := Summary{
		Count: w.count,
		Mean:  w.mean,
		Min:   w.min,
		Max:   w.max,
	}
	if w.count > 1 {
		s.StdDev = math.Sqrt(w.m2 / float64(w.count-1))
	}
	return s, true
}
