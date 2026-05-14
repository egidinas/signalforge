package controlobserve

import (
	"math"
	"time"
)

// AlgorithmBasis names the observation algorithm behind a characterization.
type AlgorithmBasis string

const (
	AlgorithmStepResponse AlgorithmBasis = "step_response_observation"
)

// TransitionCharacterization summarizes one observed control-loop response.
type TransitionCharacterization struct {
	SystemID       string
	SignalID       string
	SetpointDelta  float64
	PeakDelta      float64
	SettlingTime   time.Duration
	ControlEffort  float64
	Confidence     float64
	AlgorithmBasis AlgorithmBasis
}

// PIDAdvisorConfig tunes the conservative recommendation heuristic.
type PIDAdvisorConfig struct {
	MinimumEvents       int
	OvershootThreshold  float64
	SettlingThreshold   time.Duration
	MinorAdjustmentGain float64
}

// PIDAction is the authority level of a recommendation.
type PIDAction string

const (
	PIDActionRecommendOnly PIDAction = "recommend_only"
)

// PIDSafetyLevel describes the expected downstream review gate.
type PIDSafetyLevel string

const (
	PIDSafetyOperatorReview PIDSafetyLevel = "operator_review"
)

// PIDScaleSuggestion describes relative scale factors for PID terms.
type PIDScaleSuggestion struct {
	KpScale float64
	KiScale float64
	KdScale float64
}

// PIDRecommendation is an evidence-backed, read-only suggestion.
type PIDRecommendation struct {
	SystemID       string
	SignalID       string
	Action         PIDAction
	Safety         PIDSafetyLevel
	Suggested      PIDScaleSuggestion
	Confidence     float64
	AlgorithmBasis AlgorithmBasis
	Reasons        []string
}

// PIDAdvisor returns conservative PID scale recommendations from observations.
type PIDAdvisor struct {
	cfg PIDAdvisorConfig
}

// NewPIDAdvisor returns an advisor with conservative defaults filled.
func NewPIDAdvisor(cfg PIDAdvisorConfig) *PIDAdvisor {
	if cfg.MinimumEvents <= 0 {
		cfg.MinimumEvents = 3
	}
	if cfg.OvershootThreshold <= 0 {
		cfg.OvershootThreshold = 0.2
	}
	if cfg.MinorAdjustmentGain <= 0 {
		cfg.MinorAdjustmentGain = 0.05
	}
	return &PIDAdvisor{cfg: cfg}
}

// Recommend converts repeated transition characterizations into a suggestion.
func (a *PIDAdvisor) Recommend(events []TransitionCharacterization) (PIDRecommendation, bool) {
	if len(events) < a.cfg.MinimumEvents {
		return PIDRecommendation{}, false
	}
	rec := PIDRecommendation{
		SystemID:       events[0].SystemID,
		SignalID:       events[0].SignalID,
		Action:         PIDActionRecommendOnly,
		Safety:         PIDSafetyOperatorReview,
		Suggested:      PIDScaleSuggestion{KpScale: 1, KiScale: 1, KdScale: 1},
		AlgorithmBasis: AlgorithmStepResponse,
	}

	var confidence float64
	var overshootCount int
	var slowCount int
	for _, event := range events {
		confidence += event.Confidence
		if event.AlgorithmBasis != "" {
			rec.AlgorithmBasis = event.AlgorithmBasis
		}
		if math.Abs(event.SetpointDelta) > 1e-9 {
			overshoot := (math.Abs(event.PeakDelta) - math.Abs(event.SetpointDelta)) / math.Abs(event.SetpointDelta)
			if overshoot > a.cfg.OvershootThreshold {
				overshootCount++
			}
		}
		if a.cfg.SettlingThreshold > 0 && event.SettlingTime > a.cfg.SettlingThreshold {
			slowCount++
		}
	}
	rec.Confidence = confidence / float64(len(events))
	if overshootCount > 0 {
		rec.Suggested.KpScale = 1 - a.cfg.MinorAdjustmentGain
		rec.Suggested.KiScale = 1 - a.cfg.MinorAdjustmentGain
		rec.Reasons = append(rec.Reasons, "repeated overshoot observed after setpoint transitions")
	}
	if slowCount > 0 {
		rec.Reasons = append(rec.Reasons, "settling time exceeded observation threshold")
	}
	if len(rec.Reasons) == 0 {
		rec.Reasons = append(rec.Reasons, "behavior is stable enough for continued observation")
	}
	return rec, true
}
