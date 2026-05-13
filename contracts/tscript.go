// Package tscript defines the Loom test-script (tscript) format and executor.
//
// A Script is a named, ordered sequence of Steps. Steps can send CAN frames
// (raw or DBC-encoded), publish NATS commands with optional ACK, wait for
// durations, or assert telemetry stability. Each step reports progress to a
// NATS subject so callers can stream execution state.
//
// Subject conventions:
//
//	Submit one-shot:  req.v4.tscript.<nodeID>.run        (request → reply = ScriptResult)
//	Fire-and-forget:  control.v4.tscript.<nodeID>.load   (no reply, progress streamed)
//	Cancel:           control.v4.tscript.<nodeID>.cancel (payload: {"script_id":"..."})
//	Status:           req.v4.tscript.<nodeID>.status     (reply = RunStatus)
//	Progress stream:  event.v4.tscript.<nodeID>.<scriptID>.progress
//
// CAN frame injection (used by send_can step and loom-ctl caninject):
//
//	control.v4.cmd.hardware.can.<busID>    payload: CANWriteRequest JSON
package contracts

import (
	"time"
)

// StepKind enumerates the kinds of steps a script can contain.
type StepKind string

const (
	StepSendCAN     StepKind = "send_can"     // inject a CAN frame
	StepSendCommand StepKind = "send_command" // publish JSON to a NATS subject
	StepWait        StepKind = "wait"         // fixed sleep
	StepWaitStable  StepKind = "wait_stable"  // block until telemetry is stable
	StepAssert      StepKind = "assert"       // fail script if condition not met
	StepLog         StepKind = "log"          // emit a progress log line
)

// Step is one instruction in a Script.
type Step struct {
	ID   string   `json:"id,omitempty"` // optional human label
	Kind StepKind `json:"kind"`

	// ── send_can ─────────────────────────────────────────────────────────────
	Bus     string             `json:"bus,omitempty"`      // e.g. "ch0", "ch1"
	FrameID uint32             `json:"frame_id,omitempty"` // 11 or 29-bit CAN ID
	DLC     uint8              `json:"dlc,omitempty"`      // 0 = infer from data length
	Data    string             `json:"data,omitempty"`     // hex string e.g. "DEADBEEF01020304"
	Signals map[string]float64 `json:"signals,omitempty"`  // DBC signal name → value
	DBCPath string             `json:"dbc_path,omitempty"` // override DBC for this step

	// ── send_command ─────────────────────────────────────────────────────────
	Subject     string `json:"subject,omitempty"`
	Payload     string `json:"payload,omitempty"` // raw JSON string
	AwaitACK    bool   `json:"await_ack,omitempty"`
	ACKTimeoutS int    `json:"ack_timeout_s,omitempty"`

	// ── wait ─────────────────────────────────────────────────────────────────
	DurationMS int `json:"duration_ms,omitempty"`

	// ── wait_stable ──────────────────────────────────────────────────────────
	MonitorID      string  `json:"monitor_id,omitempty"`
	Tolerance      float64 `json:"tolerance,omitempty"`
	StableSecs     int     `json:"stable_s,omitempty"`
	StableTimeoutS int     `json:"stable_timeout_s,omitempty"`

	// ── assert ───────────────────────────────────────────────────────────────
	AssertMonitorID string  `json:"assert_monitor_id,omitempty"`
	AssertMin       float64 `json:"assert_min,omitempty"`
	AssertMax       float64 `json:"assert_max,omitempty"`

	// ── log ──────────────────────────────────────────────────────────────────
	Message string `json:"message,omitempty"`
}

// Script is the top-level tscript document.
type Script struct {
	ID       string `json:"id"`
	Label    string `json:"label,omitempty"`
	TimeoutS int    `json:"timeout_s,omitempty"` // 0 = no script-level timeout
	DBCPath  string `json:"dbc_path,omitempty"`  // default DBC for all send_can steps
	Steps    []Step `json:"steps"`
}

// StepStatus is the per-step execution outcome.
type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepRunning StepStatus = "running"
	StepOK      StepStatus = "ok"
	StepSkipped StepStatus = "skipped"
	StepFailed  StepStatus = "failed"
)

// StepResult is emitted for each completed step.
type StepResult struct {
	Index      int        `json:"index"`
	StepID     string     `json:"step_id,omitempty"`
	Kind       StepKind   `json:"kind"`
	Status     StepStatus `json:"status"`
	DurationMS int64      `json:"duration_ms"`
	Error      string     `json:"error,omitempty"`
	Note       string     `json:"note,omitempty"`
}

// ProgressEvent is published on event.v4.tscript.<nodeID>.<scriptID>.progress.
type ProgressEvent struct {
	SchemaVersion int        `json:"schema_version"`
	ScriptID      string     `json:"script_id"`
	StepIndex     int        `json:"step_index"`
	StepsTotal    int        `json:"steps_total"`
	Step          StepResult `json:"step"`
	RunningS      float64    `json:"running_s"`
}

// ScriptResult is the final reply to a req.v4.tscript.*.run request.
type ScriptResult struct {
	SchemaVersion int          `json:"schema_version"`
	ScriptID      string       `json:"script_id"`
	Status        string       `json:"status"` // "ok", "failed", "cancelled", "timeout"
	StepsOK       int          `json:"steps_ok"`
	StepsFailed   int          `json:"steps_failed"`
	DurationMS    int64        `json:"duration_ms"`
	Steps         []StepResult `json:"steps"`
	Error         string       `json:"error,omitempty"`
}

// RunStatus is the reply to req.v4.tscript.*.status.
type RunStatus struct {
	SchemaVersion int       `json:"schema_version"`
	ScriptID      string    `json:"script_id,omitempty"`
	Running       bool      `json:"running"`
	StepIndex     int       `json:"step_index"`
	StepsTotal    int       `json:"steps_total"`
	StartedAt     time.Time `json:"started_at,omitempty"`
}

// CANWriteRequest is the payload for control.v4.cmd.hardware.can.<busID>.
// Either raw Data (hex) or Signals (DBC-encoded) must be provided.
type CANWriteRequest struct {
	SchemaVersion int                `json:"schema_version"`
	FrameID       uint32             `json:"frame_id"`
	DLC           uint8              `json:"dlc,omitempty"`
	Data          string             `json:"data,omitempty"`    // hex bytes
	Signals       map[string]float64 `json:"signals,omitempty"` // DBC signal → value
	DBCPath       string             `json:"dbc_path,omitempty"`
	OriginScript  string             `json:"origin_script,omitempty"`
	OriginStep    string             `json:"origin_step,omitempty"`
}

// CANWriteSubject returns the NATS subject for CAN frame injection on a bus.
func CANWriteSubject(busID string) string {
	return "control.v4.cmd.hardware.can." + busID
}
