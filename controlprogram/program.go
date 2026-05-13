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
	return clone
}
