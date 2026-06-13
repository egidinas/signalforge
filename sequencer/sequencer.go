package sequencer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egidinas/signalforge/contracts"
)

type StepKind string

const (
	StepSendCommand StepKind = "send_command"
	StepWait        StepKind = "wait"
	StepWaitStable  StepKind = "wait_stable"
	StepAssert      StepKind = "assert"
	StepLog         StepKind = "log"
)

type Script struct {
	ID      string        `json:"id"`
	Name    string        `json:"name,omitempty"`
	Steps   []Step        `json:"steps"`
	Timeout time.Duration `json:"timeout,omitempty"`
}

type Step struct {
	ID             string            `json:"id"`
	Kind           StepKind          `json:"kind"`
	TargetID       string            `json:"target_id,omitempty"`
	CommandName    string            `json:"command_name,omitempty"`
	Arguments      map[string]any    `json:"arguments,omitempty"`
	AwaitAck       bool              `json:"await_ack,omitempty"`
	Duration       time.Duration     `json:"duration,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type StepResult struct {
	StepID   string        `json:"step_id"`
	OK       bool          `json:"ok"`
	Status   string        `json:"status,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
	Error    string        `json:"error,omitempty"`
	Note     string        `json:"note,omitempty"`
}

type Result struct {
	ScriptID string        `json:"script_id"`
	RunID    string        `json:"run_id"`
	OK       bool          `json:"ok"`
	Duration time.Duration `json:"duration,omitempty"`
	Steps    []StepResult  `json:"steps"`
	Error    string        `json:"error,omitempty"`
}

// UnmarshalJSON lets Script.timeout accept either a Go duration string or a numeric nanosecond value.
func (s *Script) UnmarshalJSON(data []byte) error {
	type alias Script
	aux := &struct {
		Timeout any `json:"timeout,omitempty"`
		*alias
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	d, err := parseFlexibleDuration(aux.Timeout)
	if err != nil {
		return fmt.Errorf("script.timeout: %w", err)
	}
	s.Timeout = d
	return nil
}

// UnmarshalJSON lets Step.duration accept either a Go duration string or a numeric nanosecond value.
func (s *Step) UnmarshalJSON(data []byte) error {
	type alias Step
	aux := &struct {
		Duration any `json:"duration,omitempty"`
		*alias
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	d, err := parseFlexibleDuration(aux.Duration)
	if err != nil {
		return fmt.Errorf("step.duration: %w", err)
	}
	s.Duration = d
	return nil
}

func parseFlexibleDuration(v any) (time.Duration, error) {
	switch x := v.(type) {
	case nil:
		return 0, nil
	case string:
		if x == "" {
			return 0, nil
		}
		return time.ParseDuration(x)
	case float64:
		return time.Duration(x), nil
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return 0, err
		}
		return time.Duration(n), nil
	default:
		return 0, fmt.Errorf("unsupported duration %T", v)
	}
}

// Handler is implemented by applications to define how specific steps are executed.
type Handler interface {
	ExecuteStep(ctx context.Context, step Step) StepResult
}

// Runner executes a script using the provided handler.
type Runner struct {
	Handler Handler
}

func (r *Runner) Run(ctx context.Context, script Script) (Result, error) {
	start := time.Now()
	runID := fmt.Sprintf("%s-run-%d", script.ID, start.UnixNano())
	res := Result{
		ScriptID: script.ID,
		RunID:    runID,
		OK:       true,
	}

	scriptCtx := ctx
	if script.Timeout > 0 {
		var cancel context.CancelFunc
		scriptCtx, cancel = context.WithTimeout(ctx, script.Timeout)
		defer cancel()
	}

	for _, step := range script.Steps {
		if err := scriptCtx.Err(); err != nil {
			res.Steps = append(res.Steps, StepResult{StepID: step.ID, OK: false, Error: err.Error()})
			res.OK = false
			break
		}

		stepStart := time.Now()
		sr := r.Handler.ExecuteStep(scriptCtx, step)
		sr.Duration = time.Since(stepStart)

		res.Steps = append(res.Steps, sr)
		if !sr.OK {
			res.OK = false
			break
		}
	}

	res.Duration = time.Since(start)
	return res, nil
}

// CommanderHandler implements Handler by sending KindSendCommand steps to a contracts.Commander.
type CommanderHandler struct {
	Commander contracts.Commander
}

func (h *CommanderHandler) ExecuteStep(ctx context.Context, step Step) StepResult {
	switch step.Kind {
	case StepSendCommand:
		tc := contracts.Telecommand{
			TargetID:       step.TargetID,
			Name:           step.CommandName,
			Arguments:      step.Arguments,
			RequiresAck:    step.AwaitAck,
			IdempotencyKey: step.IdempotencyKey,
			Time:           time.Now(),
			Metadata:       step.Metadata,
		}
		tc.EnsureIdempotencyKey()

		var ev contracts.CommandEvent
		var err error
		if ctxCmdr, ok := h.Commander.(interface {
			SendContext(context.Context, contracts.Telecommand) (contracts.CommandEvent, error)
		}); ok {
			ev, err = ctxCmdr.SendContext(ctx, tc)
		} else {
			// fallback for commanders that don't support context natively
			// wait for completion or context cancellation
			done := make(chan struct{})
			go func() {
				ev, err = h.Commander.Send(tc)
				close(done)
			}()
			select {
			case <-ctx.Done():
				return StepResult{StepID: step.ID, OK: false, Error: ctx.Err().Error()}
			case <-done:
			}
		}
		if err != nil {
			return StepResult{StepID: step.ID, OK: false, Error: err.Error()}
		}
		if ev.Status == contracts.CommandRejected || ev.Status == contracts.CommandFailed {
			msg := ev.Error
			if msg == "" {
				msg = string(ev.Status)
			}
			return StepResult{StepID: step.ID, OK: false, Error: msg}
		}
		return StepResult{StepID: step.ID, OK: true, Status: string(ev.Status)}

	case StepWait, StepWaitStable:
		if step.Duration <= 0 {
			return StepResult{StepID: step.ID, OK: true}
		}
		select {
		case <-ctx.Done():
			return StepResult{StepID: step.ID, OK: false, Error: ctx.Err().Error()}
		case <-time.After(step.Duration):
			return StepResult{StepID: step.ID, OK: true}
		}

	case StepAssert, StepLog:
		return StepResult{StepID: step.ID, OK: true}

	default:
		return StepResult{StepID: step.ID, OK: false, Error: fmt.Sprintf("unsupported step kind %q", step.Kind)}
	}
}

// Run is a helper that executes a script against a commander using the default CommanderHandler.
func Run(ctx context.Context, script Script, commander contracts.Commander) (Result, error) {
	runner := Runner{
		Handler: &CommanderHandler{Commander: commander},
	}
	return runner.Run(ctx, script)
}
