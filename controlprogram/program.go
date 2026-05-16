package controlprogram

import (
	"errors"
	"fmt"
	"time"
)

type Setpoint struct {
	Channel  string
	Value    float64
	Unit     string
	Metadata map[string]string
}

type Step struct {
	ID        string
	Hold      time.Duration
	Setpoints []Setpoint
	Metadata  map[string]string
}

type Program struct {
	ID         string
	Name       string
	TargetIDs  []string
	CycleCount int
	Steps      []Step
	Metadata   map[string]string
}

func (p Program) Validate() error {
	if p.ID == "" {
		return errors.New("program id is required")
	}
	if len(p.TargetIDs) == 0 {
		return fmt.Errorf("program %q requires at least one target", p.ID)
	}
	if p.CycleCount <= 0 {
		return fmt.Errorf("program %q requires a positive cycle count", p.ID)
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("program %q requires at least one step", p.ID)
	}

	for i, step := range p.Steps {
		if step.ID == "" {
			return fmt.Errorf("program %q step %d requires an id", p.ID, i)
		}
		if step.Hold <= 0 {
			return fmt.Errorf("program %q step %q requires a positive hold duration", p.ID, step.ID)
		}
		if len(step.Setpoints) == 0 {
			return fmt.Errorf("program %q step %q requires at least one setpoint", p.ID, step.ID)
		}
		for j, setpoint := range step.Setpoints {
			if setpoint.Channel == "" {
				return fmt.Errorf("program %q step %q setpoint %d requires a channel", p.ID, step.ID, j)
			}
			if setpoint.Unit == "" {
				return fmt.Errorf("program %q step %q setpoint %q requires a unit", p.ID, step.ID, setpoint.Channel)
			}
		}
	}

	return nil
}

func (p Program) TotalDuration() time.Duration {
	var perCycle time.Duration
	for _, step := range p.Steps {
		perCycle += step.Hold
	}
	return perCycle * time.Duration(p.CycleCount)
}

func (p Program) TargetCount() int {
	return len(p.TargetIDs)
}

func (p Program) CloneForTargets(targetIDs []string) Program {
	clone := p
	clone.TargetIDs = append([]string(nil), targetIDs...)
	clone.Steps = cloneSteps(p.Steps)
	clone.Metadata = cloneStringMap(p.Metadata)
	return clone
}

func cloneSteps(steps []Step) []Step {
	if steps == nil {
		return nil
	}
	out := make([]Step, len(steps))
	for i, step := range steps {
		out[i] = step
		out[i].Setpoints = cloneSetpoints(step.Setpoints)
		out[i].Metadata = cloneStringMap(step.Metadata)
	}
	return out
}

func cloneSetpoints(setpoints []Setpoint) []Setpoint {
	if setpoints == nil {
		return nil
	}
	out := make([]Setpoint, len(setpoints))
	for i, setpoint := range setpoints {
		out[i] = setpoint
		out[i].Metadata = cloneStringMap(setpoint.Metadata)
	}
	return out
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
