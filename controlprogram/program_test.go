package controlprogram

import (
	"testing"
	"time"
)

func TestProgramValidationAndDuration(t *testing.T) {
	program := Program{
		ID:         "sample",
		Name:       "Sample Program",
		TargetIDs:  []string{"tec-31", "tec-32"},
		CycleCount: 4,
		Steps: []Step{
			{
				ID:   "low",
				Hold: 30 * time.Second,
				Setpoints: []Setpoint{{
					Channel: "temperature.object",
					Value:   20,
					Unit:    "degC",
				}},
			},
			{
				ID:   "high",
				Hold: 30 * time.Second,
				Setpoints: []Setpoint{{
					Channel: "temperature.object",
					Value:   25,
					Unit:    "degC",
				}},
			},
		},
	}

	if err := program.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got, want := program.TotalDuration(), 4*time.Minute; got != want {
		t.Fatalf("TotalDuration() = %s, want %s", got, want)
	}
}

func TestProgramSupportsSixteenTargetsWithoutSpecialCases(t *testing.T) {
	targets := make([]string, 16)
	for i := range targets {
		targets[i] = "tec-" + string(rune('A'+i))
	}

	program := Program{
		ID:         "fanout",
		TargetIDs:  targets,
		CycleCount: 1,
		Steps: []Step{{
			ID:   "hold",
			Hold: time.Second,
			Setpoints: []Setpoint{{
				Channel: "temperature.object",
				Value:   20,
				Unit:    "degC",
			}},
		}},
	}

	if err := program.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got := program.TargetCount(); got != 16 {
		t.Fatalf("TargetCount() = %d, want 16", got)
	}
}

func TestProgramCloneForTargetsDoesNotAliasMutableFields(t *testing.T) {
	program := Program{
		ID:         "sample",
		TargetIDs:  []string{"tec-31"},
		CycleCount: 1,
		Steps: []Step{{
			ID:       "hold",
			Hold:     time.Second,
			Metadata: map[string]string{"phase": "initial"},
			Setpoints: []Setpoint{{
				Channel:  "temperature.object",
				Value:    20,
				Unit:     "degC",
				Metadata: map[string]string{"source": "plan"},
			}},
		}},
		Metadata: map[string]string{"campaign": "fat"},
	}

	clone := program.CloneForTargets([]string{"tec-32"})
	clone.TargetIDs[0] = "tec-33"
	clone.Metadata["campaign"] = "sat"
	clone.Steps[0].ID = "changed"
	clone.Steps[0].Metadata["phase"] = "changed"
	clone.Steps[0].Setpoints[0].Value = 30
	clone.Steps[0].Setpoints[0].Metadata["source"] = "operator"

	if program.TargetIDs[0] != "tec-31" {
		t.Fatalf("original target IDs were aliased: %+v", program.TargetIDs)
	}
	if program.Metadata["campaign"] != "fat" {
		t.Fatalf("original program metadata was aliased: %+v", program.Metadata)
	}
	if program.Steps[0].ID != "hold" || program.Steps[0].Metadata["phase"] != "initial" {
		t.Fatalf("original step was aliased: %+v", program.Steps[0])
	}
	if program.Steps[0].Setpoints[0].Value != 20 || program.Steps[0].Setpoints[0].Metadata["source"] != "plan" {
		t.Fatalf("original setpoint was aliased: %+v", program.Steps[0].Setpoints[0])
	}
}

func TestProgramValidationRejectsAmbiguousPrograms(t *testing.T) {
	cases := []Program{
		{},
		{ID: "missing-target", CycleCount: 1, Steps: []Step{{ID: "s", Hold: time.Second}}},
		{ID: "missing-cycle", TargetIDs: []string{"tec-31"}, Steps: []Step{{ID: "s", Hold: time.Second}}},
		{ID: "missing-step", TargetIDs: []string{"tec-31"}, CycleCount: 1},
		{ID: "missing-setpoint", TargetIDs: []string{"tec-31"}, CycleCount: 1, Steps: []Step{{ID: "s", Hold: time.Second}}},
	}

	for _, tc := range cases {
		if err := tc.Validate(); err == nil {
			t.Fatalf("Validate() for %+v returned nil, want error", tc)
		}
	}
}
