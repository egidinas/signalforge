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
	
	// Handle Target Setpoint for deviation checks
	var targetSetpoint float64
	hasTarget := false
	if m.config.TargetSignalID != "" {
		if rb, ok := m.buffers[m.config.TargetSignalID]; ok {
			if v, ok := rb.LastValue(); ok {
				targetSetpoint = v
				hasTarget = true
			}
		}
	}

	memberPasses := make([]bool, 0, len(m.config.Signals))
	for _, signal := range m.config.Signals {
		buffer, ok := m.buffers[signal]
		if !ok {
			// Should not happen if AddSignal works correctly
			continue
		}
		windowStates := make(map[string]WindowStats, len(m.config.Windows))
		status := StatusStable
		sigPasses := true
		for _, window := range m.config.Windows {
			stats := evaluateWindow(buffer, now, window, targetSetpoint, hasTarget)
			windowStates[window.Name] = stats
			status = moreSevereStatus(status, stats.Status)
			if !stats.GatePass {
				sigPasses = false
			}
		}
		if status == StatusStable {
			stableSignals++
		}
		signalStates[signal] = SignalState{Signal: signal, Windows: windowStates, Status: status}
		memberPasses = append(memberPasses, sigPasses)
	}

	total := len(signalStates)
	overall := StatusStable
	
	// Apply Group Policy
	switch m.config.GroupPolicy.Mode {
	case PolicyAllMustPass:
		if stableSignals < total {
			overall = StatusUnstable
		}
	case PolicyAnyMustPass:
		if stableSignals == 0 {
			overall = StatusUnstable
		}
	case PolicyNofM:
		if stableSignals < m.config.GroupPolicy.RequiredStable {
			overall = StatusUnstable
		}
	case PolicySpread:
		// Implement spread logic if needed, but for now just use stable count
		if stableSignals < m.config.GroupPolicy.RequiredStable {
			overall = StatusUnstable
		}
	default:
		// Legacy default
		required := m.config.GroupPolicy.RequiredStable
		if required == 0 {
			required = total
		}
		if stableSignals < required {
			overall = StatusUnstable
		}
	}

	// Final severity check across all signals if not overall stable
	if overall != StatusStable {
		worst := StatusStable
		for _, state := range signalStates {
			worst = moreSevereStatus(worst, state.Status)
		}
		// If the worst thing found is less severe than Unstable, use it.
		// Otherwise keep Unstable (or worse).
		if statusSeverity(worst) < statusSeverity(StatusUnstable) {
			overall = worst
		} else {
			overall = moreSevereStatus(overall, worst)
		}
	}

	return State{
		Signals:         signalStates,
		OverallStatus:   overall,
		StableSignals:   stableSignals,
		TotalSignals:    total,
		EvaluatedAt:     now,
		RequiredStable:  m.config.GroupPolicy.RequiredStable,
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

func evaluateAgainstCriteria(stats WindowStats, criteria Criteria) (bool, float64) {
	pass := true
	maxRatio := 0.0

	check := func(val, limit float64) {
		if limit > 0 {
			ratio := val / limit
			if ratio > 1 {
				pass = false
			}
			if ratio > maxRatio {
				maxRatio = ratio
			}
		}
	}

	check(stats.Range, criteria.MaxAbsChange)
	check(stats.Range, criteria.MaxRangeC)
	check(stats.Span, criteria.MaxSpan)
	check(stats.StdDev, criteria.MaxStdDevC)
	check(math.Abs(stats.SlopeCPerHour), criteria.MaxSlopeCPerHour)
	if stats.DeviationC != nil {
		check(*stats.DeviationC, criteria.MaxDeviationFromTarget)
		if criteria.MaxDeviationC != nil {
			check(*stats.DeviationC, *criteria.MaxDeviationC)
		}
	}

	return pass, maxRatio
}

func evaluateWindow(buffer *RollingBuffer, now time.Time, window WindowConfig, targetSetpoint float64, hasTarget bool) WindowStats {
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
			GatePass:          false,
			WarnPass:          false,
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
	span := maxValue - minValue // Same as range for now
	_, slope := EvaluateLinearDrift(samples, 0) // Just get the slope
	
	stats := WindowStats{
		Count:          len(samples),
		Mean:           mean,
		StdDev:         std,
		Min:            minValue,
		Max:            maxValue,
		Range:          valueRange,
		Span:           span,
		SlopeCPerHour:  slope,
		Status:         StatusStable,
		FirstTimestamp: samples[0].Timestamp,
		LastTimestamp:  samples[len(samples)-1].Timestamp,
		GatePass:       true,
		WarnPass:       true,
	}

	if hasTarget {
		_, deviation := EvaluateDeviationFromTarget(samples, targetSetpoint, 0)
		stats.DeviationC = &deviation
	} else if window.Criteria.TargetC != nil {
		_, deviation := EvaluateDeviationFromTarget(samples, *window.Criteria.TargetC, 0)
		stats.DeviationC = &deviation
	}

	// Gate Checks
	var gateRatio float64
	if window.Gate.MaxSpan > 0 || window.Gate.MaxAbsChange > 0 || window.Gate.MaxSlopeCPerHour > 0 || window.Gate.MaxDeviationFromTarget > 0 {
		stats.GatePass, gateRatio = evaluateAgainstCriteria(stats, window.Gate)
	} else {
		// Legacy/Default criteria
		stats.GatePass, gateRatio = evaluateAgainstCriteria(stats, window.Criteria)
	}

	// Warn Checks
	var warnRatio float64
	if window.Warn.MaxSpan > 0 || window.Warn.MaxAbsChange > 0 || window.Warn.MaxSlopeCPerHour > 0 || window.Warn.MaxDeviationFromTarget > 0 {
		stats.WarnPass, warnRatio = evaluateAgainstCriteria(stats, window.Warn)
	}

	if !stats.GatePass {
		stats.Status = StatusBreached
	} else if !stats.WarnPass {
		stats.Status = StatusApproaching
	} else {
		// Legacy approach threshold check
		ratio := gateRatio
		if warnRatio > ratio {
			ratio = warnRatio
		}
		threshold := window.Criteria.ApproachThreshold
		if threshold > 0 && ratio >= threshold {
			stats.Status = StatusApproaching
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
