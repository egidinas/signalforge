package statemachine

import (
	"fmt"
	"strings"
	"time"
)

// StateID identifies one modeled state in a machine.
type StateID string

// TransitionID identifies one modeled transition between states.
type TransitionID string

// Observation records the state observed for a machine at an event timestamp.
type Observation struct {
	MachineID   string
	State       StateID
	Timestamp   time.Time
	ObservedAt  time.Time
	EvidenceRef string
	Metadata    map[string]string
}

// Transition records one observed state change.
type Transition struct {
	MachineID   string
	ID          TransitionID
	From        StateID
	To          StateID
	Timestamp   time.Time
	ObservedAt  time.Time
	EvidenceRef string
	Metadata    map[string]string
}

// TransitionRule describes one configured state change.
type TransitionRule struct {
	ID       TransitionID
	From     StateID
	To       StateID
	Summary  string
	Metadata map[string]string
}

// Model describes a generic finite-state transition surface.
type Model struct {
	ID          string
	States      []StateID
	Transitions []TransitionRule
}

// CompiledModel indexes a model for repeated transition lookups.
type CompiledModel struct {
	ID          string
	States      []StateID
	Transitions []TransitionRule

	byPair map[transitionPair]TransitionRule
	byFrom map[StateID][]TransitionRule
}

type transitionPair struct {
	from StateID
	to   StateID
}

// Compile indexes the model while preserving the first matching transition
// semantics of Model.Allowed.
func (m Model) Compile() CompiledModel {
	compiled := CompiledModel{
		ID:          m.ID,
		States:      append([]StateID(nil), m.States...),
		Transitions: copyTransitionRules(m.Transitions),
		byPair:      make(map[transitionPair]TransitionRule, len(m.Transitions)),
		byFrom:      make(map[StateID][]TransitionRule, len(m.Transitions)),
	}
	for _, transition := range compiled.Transitions {
		pair := transitionPair{from: transition.From, to: transition.To}
		indexedTransition := copyTransitionRule(transition)
		if _, exists := compiled.byPair[pair]; !exists {
			compiled.byPair[pair] = indexedTransition
		}
		compiled.byFrom[transition.From] = append(compiled.byFrom[transition.From], indexedTransition)
	}
	return compiled
}

// Allowed returns the configured transition rule for from->to.
func (m Model) Allowed(from StateID, to StateID) (TransitionRule, bool) {
	for _, transition := range m.Transitions {
		if transition.From == from && transition.To == to {
			return copyTransitionRule(transition), true
		}
	}
	return TransitionRule{}, false
}

// AllowedFrom returns configured transition rules that leave from.
func (m Model) AllowedFrom(from StateID) []TransitionRule {
	out := make([]TransitionRule, 0, len(m.Transitions))
	for _, transition := range m.Transitions {
		if transition.From == from {
			out = append(out, copyTransitionRule(transition))
		}
	}
	return out
}

// Allowed returns the configured transition rule for from->to.
func (m CompiledModel) Allowed(from StateID, to StateID) (TransitionRule, bool) {
	transition, ok := m.byPair[transitionPair{from: from, to: to}]
	return copyTransitionRule(transition), ok
}

// AllowedFrom returns configured transition rules that leave from.
func (m CompiledModel) AllowedFrom(from StateID) []TransitionRule {
	return copyTransitionRules(m.byFrom[from])
}

// Tracker emits transitions from ordered observations.
type Tracker struct {
	lastByMachine map[string]Observation
}

// NewTracker returns an empty transition tracker.
func NewTracker() *Tracker {
	return &Tracker{lastByMachine: map[string]Observation{}}
}

// Observe adds one state observation and returns a transition when the state
// changed since the previous observation for the same machine.
func (t *Tracker) Observe(observation Observation) (Transition, bool, error) {
	if t == nil {
		return Transition{}, false, fmt.Errorf("tracker is nil")
	}
	if t.lastByMachine == nil {
		t.lastByMachine = map[string]Observation{}
	}
	if err := observation.Validate(); err != nil {
		return Transition{}, false, err
	}
	previous, ok := t.lastByMachine[observation.MachineID]
	if ok && observation.Timestamp.Before(previous.Timestamp) {
		return Transition{}, false, fmt.Errorf("observation for %q is out of order: %s before %s", observation.MachineID, observation.Timestamp.Format(time.RFC3339Nano), previous.Timestamp.Format(time.RFC3339Nano))
	}
	t.lastByMachine[observation.MachineID] = observation
	if !ok || previous.State == observation.State {
		return Transition{}, false, nil
	}
	observedAt := observation.ObservedAt
	if observedAt.IsZero() {
		observedAt = observation.Timestamp
	}
	return Transition{
		MachineID:   observation.MachineID,
		From:        previous.State,
		To:          observation.State,
		Timestamp:   observation.Timestamp,
		ObservedAt:  observedAt,
		EvidenceRef: observation.EvidenceRef,
		Metadata:    copyStringMap(observation.Metadata),
	}, true, nil
}

// ObserveWithModel adds one observation and annotates emitted transitions with
// the configured transition ID when the model allows the observed state change.
func (t *Tracker) ObserveWithModel(model Model, observation Observation) (Transition, bool, error) {
	return t.observeWithRules(model, observation)
}

// ObserveWithCompiledModel adds one observation and annotates emitted
// transitions using indexed model lookups.
func (t *Tracker) ObserveWithCompiledModel(model CompiledModel, observation Observation) (Transition, bool, error) {
	return t.observeWithRules(model, observation)
}

func (t *Tracker) observeWithRules(model interface {
	Allowed(StateID, StateID) (TransitionRule, bool)
}, observation Observation) (Transition, bool, error) {
	transition, ok, err := t.Observe(observation)
	if err != nil || !ok {
		return transition, ok, err
	}
	if rule, allowed := model.Allowed(transition.From, transition.To); allowed {
		transition.ID = rule.ID
	}
	return transition, true, nil
}

// Validate checks whether an observation can be used in a timeline.
func (o Observation) Validate() error {
	if strings.TrimSpace(o.MachineID) == "" {
		return fmt.Errorf("machine_id is required")
	}
	if strings.TrimSpace(string(o.State)) == "" {
		return fmt.Errorf("state is required")
	}
	if o.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func copyTransitionRule(in TransitionRule) TransitionRule {
	in.Metadata = copyStringMap(in.Metadata)
	return in
}

func copyTransitionRules(in []TransitionRule) []TransitionRule {
	if len(in) == 0 {
		return nil
	}
	out := make([]TransitionRule, len(in))
	for index, transition := range in {
		out[index] = copyTransitionRule(transition)
	}
	return out
}
