package statemachine

import (
	"testing"
	"time"
)

func TestTrackerEmitsStateChangeWithExactEventTimestamp(t *testing.T) {
	tracker := NewTracker()
	first := time.Date(2026, 6, 21, 8, 30, 1, 123456789, time.UTC)
	second := first.Add(250 * time.Millisecond)

	if event, ok, err := tracker.Observe(Observation{
		MachineID: "thermal-loop",
		State:     "standby",
		Timestamp: first,
	}); err != nil || ok {
		t.Fatalf("first observation = event %+v ok %t err %v, want initialization only", event, ok, err)
	}

	event, ok, err := tracker.Observe(Observation{
		MachineID:   "thermal-loop",
		State:       "active",
		Timestamp:   second,
		ObservedAt:  second.Add(3 * time.Second),
		EvidenceRef: "edge://sample/42",
	})
	if err != nil {
		t.Fatalf("observe transition: %v", err)
	}
	if !ok {
		t.Fatalf("transition not emitted")
	}
	if event.From != "standby" || event.To != "active" {
		t.Fatalf("transition = %s -> %s, want standby -> active", event.From, event.To)
	}
	if !event.Timestamp.Equal(second) {
		t.Fatalf("event timestamp = %s, want exact %s", event.Timestamp.Format(time.RFC3339Nano), second.Format(time.RFC3339Nano))
	}
	if event.Timestamp.Nanosecond() != 373456789 {
		t.Fatalf("event timestamp lost nanosecond precision: %s", event.Timestamp.Format(time.RFC3339Nano))
	}
	if event.ObservedAt.IsZero() || !event.ObservedAt.Equal(second.Add(3*time.Second)) {
		t.Fatalf("observed_at = %s, want request observed_at", event.ObservedAt.Format(time.RFC3339Nano))
	}
	if event.EvidenceRef != "edge://sample/42" {
		t.Fatalf("evidence_ref = %q", event.EvidenceRef)
	}
}

func TestTrackerRejectsOutOfOrderObservation(t *testing.T) {
	tracker := NewTracker()
	base := time.Date(2026, 6, 21, 9, 0, 0, 0, time.UTC)
	if _, _, err := tracker.Observe(Observation{
		MachineID: "loop-a",
		State:     "armed",
		Timestamp: base,
	}); err != nil {
		t.Fatalf("initial observe: %v", err)
	}

	if _, _, err := tracker.Observe(Observation{
		MachineID: "loop-a",
		State:     "ready",
		Timestamp: base.Add(-time.Nanosecond),
	}); err == nil {
		t.Fatalf("out-of-order observation accepted")
	}
}

func TestModelAllowsOnlyConfiguredTransitions(t *testing.T) {
	model := Model{
		ID: "generic-loop",
		Transitions: []TransitionRule{
			{ID: "arm", From: "idle", To: "armed"},
			{ID: "start", From: "armed", To: "active"},
		},
	}

	if rule, ok := model.Allowed("idle", "armed"); !ok || rule.ID != "arm" {
		t.Fatalf("idle->armed rule = %+v ok %t", rule, ok)
	}
	if _, ok := model.Allowed("active", "idle"); ok {
		t.Fatalf("unconfigured active->idle transition was allowed")
	}
	if allowed := model.AllowedFrom("armed"); len(allowed) != 1 || allowed[0].ID != "start" {
		t.Fatalf("allowed from armed = %+v", allowed)
	}
}

func TestCompiledModelPreservesLookupSemantics(t *testing.T) {
	model := Model{
		ID: "generic-loop",
		Transitions: []TransitionRule{
			{ID: "first", From: "idle", To: "armed", Metadata: map[string]string{"label": "original"}},
			{ID: "duplicate", From: "idle", To: "armed"},
			{ID: "start", From: "armed", To: "active"},
		},
	}
	compiled := model.Compile()
	model.Transitions[0].Metadata["label"] = "source-mutated"

	if compiled.ID != model.ID {
		t.Fatalf("compiled ID = %q, want %q", compiled.ID, model.ID)
	}
	if rule, ok := compiled.Allowed("idle", "armed"); !ok || rule.ID != "first" {
		t.Fatalf("compiled idle->armed rule = %+v ok %t, want first match", rule, ok)
	} else {
		rule.Metadata["label"] = "allowed-mutated"
	}
	allowed := compiled.AllowedFrom("idle")
	if len(allowed) != 2 || allowed[0].ID != "first" || allowed[1].ID != "duplicate" {
		t.Fatalf("compiled allowed from idle = %+v, want original order", allowed)
	}
	allowed[0].ID = "mutated"
	allowed[0].Metadata["label"] = "allowed-from-mutated"
	compiled.Transitions[0].Metadata["label"] = "compiled-snapshot-mutated"
	if again := compiled.AllowedFrom("idle"); again[0].ID != "first" {
		t.Fatalf("compiled AllowedFrom exposed internal slice: %+v", again)
	} else if again[0].Metadata["label"] != "original" {
		t.Fatalf("compiled AllowedFrom exposed metadata alias: %+v", again[0].Metadata)
	}
	if rule, ok := compiled.Allowed("idle", "armed"); !ok || rule.Metadata["label"] != "original" {
		t.Fatalf("compiled Allowed metadata = %+v ok %t, want original", rule.Metadata, ok)
	}
}

func TestTrackerUsesCompiledModelForTransitionIDs(t *testing.T) {
	tracker := NewTracker()
	model := Model{
		ID: "generic-loop",
		Transitions: []TransitionRule{
			{ID: "start", From: "armed", To: "active"},
		},
	}.Compile()
	base := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	if _, _, err := tracker.ObserveWithCompiledModel(model, Observation{
		MachineID: "loop-a",
		State:     "armed",
		Timestamp: base,
	}); err != nil {
		t.Fatalf("initial observe: %v", err)
	}
	event, ok, err := tracker.ObserveWithCompiledModel(model, Observation{
		MachineID: "loop-a",
		State:     "active",
		Timestamp: base.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("observe transition: %v", err)
	}
	if !ok || event.ID != "start" {
		t.Fatalf("compiled transition = %+v ok %t, want start transition", event, ok)
	}
}
