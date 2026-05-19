package stability

import "time"

type StabilityStatus string

const (
	StatusInsufficientHistory StabilityStatus = "insufficient_history"
	StatusUnstable            StabilityStatus = "unstable"
	StatusApproaching         StabilityStatus = "approaching"
	StatusStable              StabilityStatus = "stable"
	StatusBreached            StabilityStatus = "breached"
)

type Criteria struct {
	MaxStdDevC             float64            `json:"max_stddev_c,omitempty"`
	MaxRangeC              float64            `json:"max_range_c,omitempty"`
	MaxAbsChange           float64            `json:"max_abs_change,omitempty"`
	MaxSpan                float64            `json:"max_span,omitempty"`
	MaxSlopeCPerHour       float64            `json:"max_slope_c_per_hour,omitempty"`
	TargetC                *float64           `json:"target_c,omitempty"`
	MaxDeviationC          *float64           `json:"max_deviation_c,omitempty"`
	MaxDeviationFromTarget float64            `json:"max_deviation_from_target,omitempty"`
	Metadata               map[string]string  `json:"metadata,omitempty"`
	BreachThresholds       map[string]float64 `json:"breach_thresholds,omitempty"`
	ApproachThreshold      float64            `json:"approach_threshold,omitempty"`
}

type WindowConfig struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Warn     Criteria      `json:"warn,omitempty"`
	Gate     Criteria      `json:"gate,omitempty"`
	Criteria Criteria      `json:"criteria,omitempty"` // Legacy/Default criteria
}

type GroupPolicyMode string

const (
	PolicyAllMustPass GroupPolicyMode = "all_must_pass"
	PolicyAnyMustPass GroupPolicyMode = "any_must_pass"
	PolicySpread      GroupPolicyMode = "spread"
	PolicyNofM        GroupPolicyMode = "n_of_m"
)

type GroupPolicy struct {
	Mode           GroupPolicyMode `json:"mode"`
	RequiredStable int             `json:"required_stable"`
	TotalSignals   int             `json:"total_signals"`
	MaxGroupSpread float64         `json:"max_group_spread,omitempty"`
}

type Config struct {
	ID             string         `json:"id"`
	Description    string         `json:"description,omitempty"`
	Signals        []string       `json:"signals"`
	TargetSignalID string         `json:"target_signal_id,omitempty"`
	Windows        []WindowConfig `json:"windows"`
	GroupPolicy    GroupPolicy    `json:"group_policy"`
}

type WindowStats struct {
	Count             int             `json:"count"`
	Mean              float64         `json:"mean"`
	StdDev            float64         `json:"stddev"`
	Min               float64         `json:"min"`
	Max               float64         `json:"max"`
	Range             float64         `json:"range"`
	Span              float64         `json:"span"`
	SlopeCPerHour     float64         `json:"slope_c_per_hour"`
	DeviationC        *float64        `json:"deviation_c,omitempty"`
	Status            StabilityStatus `json:"status"`
	GatePass          bool            `json:"gate_pass"`
	WarnPass          bool            `json:"warn_pass"`
	FirstTimestamp    time.Time       `json:"first_timestamp"`
	LastTimestamp     time.Time       `json:"last_timestamp"`
	InsufficientUntil *time.Time      `json:"insufficient_until,omitempty"`
	BreachReason      string          `json:"breach_reason,omitempty"`
}

type SignalState struct {
	Signal  string                 `json:"signal"`
	Windows map[string]WindowStats `json:"windows"`
	Status  StabilityStatus        `json:"status"`
}

type State struct {
	Signals         map[string]SignalState `json:"signals"`
	OverallStatus   StabilityStatus        `json:"overall_status"`
	StableSignals   int                    `json:"stable_signals"`
	TotalSignals    int                    `json:"total_signals"`
	EvaluatedAt     time.Time              `json:"evaluated_at"`
	RequiredStable  int                    `json:"required_stable"`
	RequiredSignals int                    `json:"required_signals"`
}

type DwellGateStateJSON struct {
	StateStart                 time.Time       `json:"state_start"`
	LastEvaluation             time.Time       `json:"last_evaluation"`
	AccumulatedAdherence       time.Duration   `json:"accumulated_adherence"`
	LongestContinuousAdherence time.Duration   `json:"longest_continuous_adherence"`
	CurrentContinuousStart     *time.Time      `json:"current_continuous_start,omitempty"`
	Satisfied                  bool            `json:"satisfied"`
	Status                     StabilityStatus `json:"status"`
}

type BreachPolicy struct {
	ResetOnBreach bool `json:"reset_on_breach"`
	PauseOnBreach bool `json:"pause_on_breach"`
}

type DwellGateState struct {
	StateStart                 time.Time
	LastEvaluation             time.Time
	AccumulatedAdherence       time.Duration
	LongestContinuousAdherence time.Duration
	CurrentContinuousStart     *time.Time
	Satisfied                  bool
	Status                     StabilityStatus
}

func (d DwellGateState) ToJSON() DwellGateStateJSON {
	return DwellGateStateJSON{
		StateStart:                 d.StateStart,
		LastEvaluation:             d.LastEvaluation,
		AccumulatedAdherence:       d.AccumulatedAdherence,
		LongestContinuousAdherence: d.LongestContinuousAdherence,
		CurrentContinuousStart:     d.CurrentContinuousStart,
		Satisfied:                  d.Satisfied,
		Status:                     d.Status,
	}
}

type StabilityEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Signal    string    `json:"signal"`
	Value     float64   `json:"value"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Template struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Config      Config       `json:"config"`
	Tags        []string     `json:"tags,omitempty"`
	Windows     []WindowHint `json:"windows,omitempty"`
}

type WindowHint struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Monitor struct {
	config  Config
	buffers map[string]*RollingBuffer
}

type point struct {
	t time.Time
	v float64
}

type RollingBuffer struct {
	horizon time.Duration
	points  []point
}

func NewRollingBuffer(horizon time.Duration) *RollingBuffer {
	return &RollingBuffer{horizon: horizon}
}

func (rb *RollingBuffer) Push(ts time.Time, value float64) {
	rb.points = append(rb.points, point{t: ts, v: value})
	cutoff := ts.Add(-rb.horizon)
	idx := 0
	for idx < len(rb.points) && rb.points[idx].t.Before(cutoff) {
		idx++
	}
	if idx > 0 {
		rb.points = append([]point(nil), rb.points[idx:]...)
	}
}

func (rb *RollingBuffer) Snapshot() []StabilityEvent {
	events := make([]StabilityEvent, 0, len(rb.points))
	for _, p := range rb.points {
		events = append(events, StabilityEvent{Timestamp: p.t, Value: p.v})
	}
	return events
}

func (rb *RollingBuffer) LastValue() (float64, bool) {
	if len(rb.points) == 0 {
		return 0, false
	}
	return rb.points[len(rb.points)-1].v, true
}

func NewMonitor(config Config) *Monitor {
	maxHorizon := time.Duration(0)
	for _, w := range config.Windows {
		if w.Duration > maxHorizon {
			maxHorizon = w.Duration
		}
	}
	buffers := make(map[string]*RollingBuffer, len(config.Signals))
	for _, signal := range config.Signals {
		buffers[signal] = NewRollingBuffer(maxHorizon)
	}
	if config.TargetSignalID != "" {
		if _, ok := buffers[config.TargetSignalID]; !ok {
			buffers[config.TargetSignalID] = NewRollingBuffer(maxHorizon)
		}
	}
	return &Monitor{config: config, buffers: buffers}
}
