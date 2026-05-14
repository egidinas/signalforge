package controlobserve

import (
	"math"
	"time"
)

// Sample is one scalar observation from a control loop.
type Sample struct {
	Time     time.Time
	SystemID string
	SignalID string
	Value    float64
	Unit     string
}

// TransitionDirection describes the first observed direction of a transition.
type TransitionDirection string

const (
	TransitionRising      TransitionDirection = "rising"
	TransitionFalling     TransitionDirection = "falling"
	TransitionDisturbance TransitionDirection = "disturbance"
)

// TransitionEvent is emitted when the detector sees a value leave baseline.
type TransitionEvent struct {
	SystemID   string
	SignalID   string
	Start      time.Time
	Direction  TransitionDirection
	Baseline   float64
	PeakDelta  float64
	NoiseSigma float64
	Confidence float64
}

// TransitionDetectorConfig tunes the EWMA baseline detector.
type TransitionDetectorConfig struct {
	Alpha            float64
	MinimumBaseline  int
	MinimumMagnitude float64
	SigmaThreshold   float64
}

// TransitionDetector detects the first significant departure from a baseline.
//
// The detector is intentionally small and deterministic. It is suitable for
// spotting candidate step responses, not for claiming process safety.
type TransitionDetector struct {
	cfg       TransitionDetectorConfig
	count     int
	mean      float64
	variance  float64
	triggered bool
}

// NewTransitionDetector returns a detector with conservative defaults filled.
func NewTransitionDetector(cfg TransitionDetectorConfig) *TransitionDetector {
	if cfg.Alpha <= 0 || cfg.Alpha >= 1 {
		cfg.Alpha = 0.1
	}
	if cfg.MinimumBaseline <= 0 {
		cfg.MinimumBaseline = 20
	}
	if cfg.SigmaThreshold <= 0 {
		cfg.SigmaThreshold = 5
	}
	return &TransitionDetector{cfg: cfg}
}

// Observe adds one sample and may return the first transition event.
func (d *TransitionDetector) Observe(sample Sample) (TransitionEvent, bool) {
	if math.IsNaN(sample.Value) {
		return TransitionEvent{}, false
	}
	if d.count == 0 {
		d.count = 1
		d.mean = sample.Value
		return TransitionEvent{}, false
	}

	sigma := math.Sqrt(math.Max(d.variance, 0))
	delta := sample.Value - d.mean
	threshold := math.Max(d.cfg.MinimumMagnitude, d.cfg.SigmaThreshold*sigma)
	if d.count >= d.cfg.MinimumBaseline && !d.triggered && math.Abs(delta) >= threshold {
		d.triggered = true
		direction := TransitionRising
		if delta < 0 {
			direction = TransitionFalling
		}
		confidence := 1.0
		if threshold > 0 {
			confidence = math.Min(1, math.Abs(delta)/math.Max(threshold*2, 1e-9))
		}
		return TransitionEvent{
			SystemID:   sample.SystemID,
			SignalID:   sample.SignalID,
			Start:      sample.Time,
			Direction:  direction,
			Baseline:   d.mean,
			PeakDelta:  math.Abs(delta),
			NoiseSigma: sigma,
			Confidence: confidence,
		}, true
	}

	alpha := d.cfg.Alpha
	nextMean := d.mean + alpha*delta
	nextDelta := sample.Value - nextMean
	d.variance = (1 - alpha) * (d.variance + alpha*delta*nextDelta)
	d.mean = nextMean
	d.count++
	return TransitionEvent{}, false
}
