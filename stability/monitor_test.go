package stability

import (
	"testing"
	"time"
)

func testConfig() Config {
	return Config{
		Signals: []string{"tec1"},
		Windows: []WindowConfig{{
			Name:     "short",
			Duration: time.Minute,
			Criteria: Criteria{
				MaxStdDevC:       0.2,
				MaxRangeC:        0.5,
				MaxSlopeCPerHour: 2,
			},
		}},
		GroupPolicy: GroupPolicy{RequiredStable: 1, TotalSignals: 1},
	}
}

func TestMonitorStartsInsufficientHistory(t *testing.T) {
	cfg := testConfig()
	monitor := NewMonitor(cfg)
	now := time.Unix(1000, 0)
	monitor.Push("tec1", now, 20.0)
	state := monitor.Evaluate(now)
	if state.OverallStatus != StatusInsufficientHistory {
		t.Fatalf("expected insufficient history, got %s", state.OverallStatus)
	}
}

func TestMonitorStableAfterFlatSignal(t *testing.T) {
	cfg := testConfig()
	monitor := NewMonitor(cfg)
	start := time.Unix(1000, 0)
	for i := 0; i <= 60; i += 10 {
		monitor.Push("tec1", start.Add(time.Duration(i)*time.Second), 20.0)
	}
	state := monitor.Evaluate(start.Add(time.Minute))
	if state.OverallStatus != StatusStable {
		t.Fatalf("expected stable, got %s", state.OverallStatus)
	}
}

func TestMonitorUnstableOnRamp(t *testing.T) {
	cfg := testConfig()
	cfg.Windows[0].Criteria.MaxRangeC = 0.2
	monitor := NewMonitor(cfg)
	start := time.Unix(1000, 0)
	for i := 0; i <= 60; i += 10 {
		monitor.Push("tec1", start.Add(time.Duration(i)*time.Second), 20.0+float64(i)/60.0)
	}
	state := monitor.Evaluate(start.Add(time.Minute))
	if state.OverallStatus == StatusStable {
		t.Fatalf("expected non-stable state for ramp")
	}
}

func TestRollingBufferEviction(t *testing.T) {
	rb := NewRollingBuffer(time.Second)
	start := time.Unix(1000, 0)
	rb.Push(start, 1)
	rb.Push(start.Add(500*time.Millisecond), 2)
	rb.Push(start.Add(1500*time.Millisecond), 3)
	snapshot := rb.Snapshot()
	if len(snapshot) != 2 {
		t.Fatalf("expected 2 retained points, got %d", len(snapshot))
	}
	if snapshot[0].Value != 2 || snapshot[1].Value != 3 {
		t.Fatalf("unexpected snapshot values: %+v", snapshot)
	}
}

func TestBufferExportImport(t *testing.T) {
	cfg := testConfig()
	source := NewMonitor(cfg)
	start := time.Unix(1000, 0)
	for i := 0; i <= 60; i += 10 {
		source.Push("tec1", start.Add(time.Duration(i)*time.Second), 20.0)
	}

	restored := NewMonitor(cfg)
	restored.SetBuffers(source.GetBuffers())
	state := restored.Evaluate(start.Add(time.Minute))
	if state.OverallStatus != StatusStable {
		t.Fatalf("expected restored buffers to evaluate stable, got %s", state.OverallStatus)
	}
}

func TestDwellTransition(t *testing.T) {
	start := time.Unix(1000, 0)
	dwell := time.Minute
	state := DwellTransition(DwellGateState{}, true, start, dwell, BreachPolicy{PauseOnBreach: true})
	if state.Status != StatusApproaching {
		t.Fatalf("expected approaching on first stable sample, got %s", state.Status)
	}
	state = DwellTransition(state, true, start.Add(dwell+time.Second), dwell, BreachPolicy{PauseOnBreach: true})
	if !state.Satisfied || state.Status != StatusStable {
		t.Fatalf("expected dwell satisfied, got %+v", state)
	}
	state = DwellTransition(state, false, start.Add(2*dwell), dwell, BreachPolicy{ResetOnBreach: true})
	if state.Satisfied || state.AccumulatedAdherence != 0 || state.Status != StatusBreached {
		t.Fatalf("expected reset breach, got %+v", state)
	}
}

func TestValidateConfig(t *testing.T) {
	errs := ValidateConfig(Config{})
	if len(errs) == 0 {
		t.Fatalf("expected validation errors")
	}
}
