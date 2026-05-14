package stability

import (
	"math"
	"time"
)

func EvaluateLinearDrift(samples []StabilityEvent, limitPerHour float64) (bool, float64) {
	slope := linearSlope(samples)
	return math.Abs(slope) <= limitPerHour, slope
}

func linearSlope(samples []StabilityEvent) float64 {
	if len(samples) < 2 {
		return 0
	}
	base := samples[0].Timestamp
	var sumX, sumY, sumXY, sumXX float64
	for _, sample := range samples {
		x := sample.Timestamp.Sub(base).Hours()
		y := sample.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}
	n := float64(len(samples))
	denom := n*sumXX - sumX*sumX
	if math.Abs(denom) < 1e-12 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

func EvaluateDeviationFromTarget(samples []StabilityEvent, target, maxDeviation float64) (bool, float64) {
	if len(samples) == 0 {
		return true, 0
	}
	last := samples[len(samples)-1].Value
	deviation := math.Abs(last - target)
	return deviation <= maxDeviation, deviation
}

func EvaluateNofMGroupPolicy(states map[string]SignalState, required int) (bool, int) {
	stable := 0
	for _, state := range states {
		if state.Status == StatusStable {
			stable++
		}
	}
	return stable >= required, stable
}

func stddev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		d := value - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(values)))
}

func finiteOrZero(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func wallClockNow() time.Time {
	return time.Now()
}
