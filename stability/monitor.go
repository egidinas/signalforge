package stability

import (
	"fmt"
	"math"
	"time"
)

func (m *Monitor) Push(signal string, ts time.Time, value float64) {
	if _, ok := m.buffers[signal]; !ok {
		m.AddSignal(signal)
	}
	m.buffers[signal].Push(ts, value)
}

func (m *Monitor) AddSignal(signal string) {
	if _, ok := m.buffers[signal]; ok {
		return
	}
	maxHorizon := time.Duration(0)
	for _, w := range m.config.Windows {
		if w.Duration > maxHorizon {
			maxHorizon = w.Duration
		}
	}
	m.config.Signals = append(m.config.Signals, signal)
	m.buffers[signal] = NewRollingBuffer(maxHorizon)
}

func (m *Monitor) Evaluate(now time.Time) State {
	signalStates := make(map[string]SignalState, len(m.buffers))
	stableSignals := 0
	for signal, buffer := range m.buffers {
		windowStates := make(map[string]WindowStats, len(m.config.Windows))
		status := StatusStable
		for _, window := range m.config.Windows {
			stats := evaluateWindow(buffer, now, window)
			windowStates[window.Name] = stats
			status = moreSevereStatus(status, stats.Status)
		}
		if status == StatusStable {
			stableSignals++
		}
		signalStates[signal] = SignalState{Signal: signal, Windows: windowStates, Status: status}
	}

	required := m.config.GroupPolicy.RequiredStable
	if required == 0 {
		required = len(m.config.Signals)
	}
	total := len(signalStates)
	overall := StatusStable
	if stableSignals < required {
		for _, state := range signalStates {
			overall = moreSevereStatus(overall, state.Status)
		}
		if overall == StatusStable {
			overall = StatusUnstable
		}
	}
	return State{
		Signals:         signalStates,
		OverallStatus:   overall,
		StableSignals:   stableSignals,
		TotalSignals:    total,
		EvaluatedAt:     now,
		RequiredStable:  required,
		RequiredSignals: m.config.GroupPolicy.TotalSignals,
	}
}

func moreSevereStatus(current, next StabilityStatus) StabilityStatus {
	if statusSeverity(next) > statusSeverity(current) {
		return next
	}
	return current
}

func statusSeverity(status StabilityStatus) int {
	switch status {
	case StatusStable:
		return 0
	case StatusApproaching:
		return 1
	case StatusUnstable:
		return 2
	case StatusInsufficientHistory:
		return 3
	case StatusBreached:
		return 4
	default:
		return 2
	}
}

func evaluateWindow(buffer *RollingBuffer, now time.Time, window WindowConfig) WindowStats {
	cutoff := now.Add(-window.Duration)
	var samples []StabilityEvent
	for _, point := range buffer.points {
		if !point.t.Before(cutoff) && !point.t.After(now) {
			samples = append(samples, StabilityEvent{Timestamp: point.t, Value: point.v})
		}
	}
	if len(samples) == 0 || buffer == nil || len(buffer.points) == 0 || buffer.points[0].t.After(cutoff) {
		insufficientUntil := cutoff.Add(window.Duration)
		return WindowStats{
			Count:             len(samples),
			Status:            StatusInsufficientHistory,
			InsufficientUntil: &insufficientUntil,
		}
	}

	values := make([]float64, len(samples))
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)
	var sum float64
	for i, sample := range samples {
		values[i] = sample.Value
		sum += sample.Value
		if sample.Value < minValue {
			minValue = sample.Value
		}
		if sample.Value > maxValue {
			maxValue = sample.Value
		}
	}
	mean := sum / float64(len(samples))
	std := finiteOrZero(stddev(values, mean))
	valueRange := maxValue - minValue
	_, slope := EvaluateLinearDrift(samples, window.Criteria.MaxSlopeCPerHour)
	stats := WindowStats{
		Count:          len(samples),
		Mean:           mean,
		StdDev:         std,
		Min:            minValue,
		Max:            maxValue,
		Range:          valueRange,
		SlopeCPerHour:  slope,
		Status:         StatusStable,
		FirstTimestamp: samples[0].Timestamp,
		LastTimestamp:  samples[len(samples)-1].Timestamp,
	}

	if window.Criteria.TargetC != nil && window.Criteria.MaxDeviationC != nil {
		_, deviation := EvaluateDeviationFromTarget(samples, *window.Criteria.TargetC, *window.Criteria.MaxDeviationC)
		stats.DeviationC = &deviation
	}

	excess, reason := computeExcess(stats, window.Criteria)
	if reason != "" {
		stats.BreachReason = reason
		if excess >= 1 {
			stats.Status = StatusBreached
		} else if window.Criteria.ApproachThreshold > 0 && excess >= window.Criteria.ApproachThreshold {
			stats.Status = StatusApproaching
		} else {
			stats.Status = StatusUnstable
		}
	}
	return stats
}

func computeExcess(stats WindowStats, criteria Criteria) (float64, string) {
	checks := []struct {
		name  string
		value float64
		limit float64
	}{
		{name: "stddev", value: stats.StdDev, limit: criteria.MaxStdDevC},
		{name: "range", value: stats.Range, limit: criteria.MaxRangeC},
		{name: "slope", value: math.Abs(stats.SlopeCPerHour), limit: criteria.MaxSlopeCPerHour},
	}
	if criteria.TargetC != nil && criteria.MaxDeviationC != nil && stats.DeviationC != nil {
		checks = append(checks, struct {
			name  string
			value float64
			limit float64
		}{name: "deviation", value: *stats.DeviationC, limit: *criteria.MaxDeviationC})
	}
	maxExcess := 0.0
	reason := ""
	for _, check := range checks {
		if check.limit <= 0 {
			continue
		}
		ratio := check.value / check.limit
		if ratio > maxExcess {
			maxExcess = ratio
			reason = fmt.Sprintf("%s %.6g approaches limit %.6g", check.name, check.value, check.limit)
		}
	}
	threshold := 1.0
	if criteria.ApproachThreshold > 0 && criteria.ApproachThreshold < 1 {
		threshold = criteria.ApproachThreshold
	}
	if maxExcess < threshold {
		return maxExcess, ""
	}
	if maxExcess >= 1 {
		reason = ""
		for _, check := range checks {
			if check.limit <= 0 {
				continue
			}
			ratio := check.value / check.limit
			if ratio == maxExcess {
				reason = fmt.Sprintf("%s %.6g exceeds limit %.6g", check.name, check.value, check.limit)
				break
			}
		}
	}
	return maxExcess, reason
}

func describeBreach(stats WindowStats, criteria Criteria) string {
	_, reason := computeExcess(stats, criteria)
	return reason
}

func (m *Monitor) GetBuffers() map[string][]StabilityEvent {
	out := make(map[string][]StabilityEvent, len(m.buffers))
	for signal, buffer := range m.buffers {
		out[signal] = buffer.Snapshot()
	}
	return out
}

func (m *Monitor) SetBuffers(buffers map[string][]StabilityEvent) {
	for signal, events := range buffers {
		if _, ok := m.buffers[signal]; !ok {
			m.AddSignal(signal)
		}
		rb := NewRollingBuffer(maxWindowHorizon(m.config.Windows))
		for _, event := range events {
			rb.Push(event.Timestamp, event.Value)
		}
		m.buffers[signal] = rb
	}
}

func maxWindowHorizon(windows []WindowConfig) time.Duration {
	maxHorizon := time.Duration(0)
	for _, window := range windows {
		if window.Duration > maxHorizon {
			maxHorizon = window.Duration
		}
	}
	return maxHorizon
}

func DwellTransition(previous DwellGateState, stable bool, now time.Time, dwell time.Duration, policy BreachPolicy) DwellGateState {
	state := previous
	if state.StateStart.IsZero() {
		state.StateStart = now
	}
	if stable {
		if state.CurrentContinuousStart == nil {
			start := now
			state.CurrentContinuousStart = &start
		}
		continuous := now.Sub(*state.CurrentContinuousStart)
		if continuous > state.LongestContinuousAdherence {
			state.LongestContinuousAdherence = continuous
		}
		if state.LastEvaluation.IsZero() {
			state.LastEvaluation = now
		}
		state.AccumulatedAdherence += now.Sub(state.LastEvaluation)
		if state.AccumulatedAdherence >= dwell || continuous >= dwell {
			state.Satisfied = true
			state.Status = StatusStable
		} else {
			state.Status = StatusApproaching
		}
	} else {
		state.CurrentContinuousStart = nil
		if policy.ResetOnBreach {
			state.AccumulatedAdherence = 0
			state.LongestContinuousAdherence = 0
			state.Satisfied = false
		}
		if policy.PauseOnBreach {
			state.Status = StatusUnstable
		} else {
			state.Status = StatusBreached
			state.Satisfied = false
		}
	}
	state.LastEvaluation = now
	return state
}

func EvaluateDwellGate(state DwellGateState, current State, dwell time.Duration, policy BreachPolicy) DwellGateState {
	stable := current.OverallStatus == StatusStable
	return DwellTransition(state, stable, current.EvaluatedAt, dwell, policy)
}

func ValidateConfig(config Config) []ValidationError {
	var errs []ValidationError
	if len(config.Signals) == 0 {
		errs = append(errs, ValidationError{Field: "signals", Message: "at least one signal is required"})
	}
	if len(config.Windows) == 0 {
		errs = append(errs, ValidationError{Field: "windows", Message: "at least one window is required"})
	}
	for i, window := range config.Windows {
		if window.Name == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("windows[%d].name", i), Message: "window name is required"})
		}
		if window.Duration <= 0 {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("windows[%d].duration", i), Message: "window duration must be positive"})
		}
		if window.Criteria.MaxStdDevC < 0 || window.Criteria.MaxRangeC < 0 || window.Criteria.MaxSlopeCPerHour < 0 {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("windows[%d].criteria", i), Message: "criteria limits must be non-negative"})
		}
	}
	if config.GroupPolicy.RequiredStable < 0 {
		errs = append(errs, ValidationError{Field: "group_policy.required_stable", Message: "required stable cannot be negative"})
	}
	return errs
}
