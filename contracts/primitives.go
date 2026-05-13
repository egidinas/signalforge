package contracts

import (
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const CommandSessionSchemaVersion = 1

const (
	EventCommandAccepted     = "tc.command_accepted"
	EventCommandDispatched   = "tc.command_dispatched"
	EventCommandRunning      = "tm.command_running"
	EventCommandProgress     = "tm.command_progress"
	EventResultRecorded      = "tm.result_recorded"
	EventCommandFailed       = "tm.command_failed"
	EventCommandTimeout      = "tm.command_timeout"
	EventControllerAttached  = "tm.controller_attached"
	EventControllerHeartbeat = "tm.controller_heartbeat"
	EventWorkerPoll          = "tm.worker_poll"
)

type TransportKind string

const (
	TransportNATSRPC      TransportKind = "nats_rpc"
	TransportLocalREST    TransportKind = "local_rest"
	TransportSupportHTTP  TransportKind = "support_http"
	TransportSSHExec      TransportKind = "ssh_exec"
	TransportLocalProcess TransportKind = "local_process"
)

type TransportAttempt struct {
	Kind         TransportKind `json:"kind"`
	Route        string        `json:"route,omitempty"`
	FallbackRank int           `json:"fallback_rank,omitempty"`
	Status       string        `json:"status,omitempty"`
	Error        string        `json:"error,omitempty"`
	StartedAt    time.Time     `json:"started_at,omitempty"`
	FinishedAt   time.Time     `json:"finished_at,omitempty"`
}

type CommandRoute struct {
	Preferred TransportKind      `json:"preferred"`
	Attempts  []TransportAttempt `json:"attempts,omitempty"`
	Winning   *TransportAttempt  `json:"winning,omitempty"`
}

type IdempotencyPolicy struct {
	Key              string `json:"key,omitempty"`
	Idempotent       bool   `json:"idempotent"`
	NonIdempotent    bool   `json:"non_idempotent,omitempty"`
	ExceptionReason  string `json:"exception_reason,omitempty"`
	ReplayProtection string `json:"replay_protection,omitempty"`
}

type CommandAck struct {
	SchemaVersion       int               `json:"schema_version"`
	Kind                string            `json:"kind"`
	CommandID           string            `json:"command_id"`
	Host                string            `json:"host"`
	Action              string            `json:"action"`
	Accepted            bool              `json:"accepted"`
	Status              string            `json:"status"`
	ControllerSessionID string            `json:"controller_session_id,omitempty"`
	ExpiresAt           time.Time         `json:"expires_at"`
	Sequence            int64             `json:"sequence,omitempty"`
	Route               *CommandRoute     `json:"route,omitempty"`
	Idempotency         IdempotencyPolicy `json:"idempotency"`
}

type ResultAck struct {
	SchemaVersion int           `json:"schema_version"`
	Kind          string        `json:"kind"`
	CommandID     string        `json:"command_id"`
	Host          string        `json:"host"`
	Status        string        `json:"status"`
	Recorded      bool          `json:"recorded"`
	ReceiptHash   string        `json:"receipt_hash,omitempty"`
	Sequence      int64         `json:"sequence,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
	Route         *CommandRoute `json:"route,omitempty"`
}

type CommandSession struct {
	ID               string    `json:"id"`
	ControllerID     string    `json:"controller_id"`
	Host             string    `json:"host,omitempty"`
	SubscriberRole   string    `json:"subscriber_role"`
	SubscribedKinds  []string  `json:"subscribed_kinds"`
	AttachMode       string    `json:"attach_mode"`
	CreatedAt        time.Time `json:"created_at"`
	LastSeen         time.Time `json:"last_seen"`
	LastDeliveredSeq int64     `json:"last_delivered_seq,omitempty"`
	Status           string    `json:"status"`
}

type Event struct {
	SchemaVersion       int             `json:"schema_version"`
	Sequence            int64           `json:"sequence"`
	Kind                string          `json:"kind"`
	Host                string          `json:"host,omitempty"`
	ControllerSessionID string          `json:"controller_session_id,omitempty"`
	CommandID           string          `json:"command_id,omitempty"`
	Status              string          `json:"status,omitempty"`
	Timestamp           time.Time       `json:"timestamp"`
	Payload             json.RawMessage `json:"payload,omitempty"`
}

func NewCommandAck(commandID, host, action, status, controllerSessionID string, expiresAt time.Time) CommandAck {
	if status == "" {
		status = "queued"
	}
	commandID = strings.TrimSpace(commandID)
	return CommandAck{
		SchemaVersion:       SchemaVersion,
		Kind:                EventCommandAccepted,
		CommandID:           commandID,
		Host:                strings.TrimSpace(host),
		Action:              strings.TrimSpace(action),
		Accepted:            true,
		Status:              strings.TrimSpace(status),
		ControllerSessionID: strings.TrimSpace(controllerSessionID),
		ExpiresAt:           expiresAt,
		Idempotency:         NewIdempotencyPolicy(commandID),
	}
}

func NewIdempotencyPolicy(key string) IdempotencyPolicy {
	key = strings.TrimSpace(key)
	return IdempotencyPolicy{
		Key:              key,
		Idempotent:       true,
		ReplayProtection: "idempotency_key",
	}
}

func NewNonIdempotentPolicy(reason string) IdempotencyPolicy {
	reason = strings.TrimSpace(reason)
	return IdempotencyPolicy{
		Idempotent:      false,
		NonIdempotent:   true,
		ExceptionReason: reason,
	}
}

func ValidateIdempotencyPolicy(policy IdempotencyPolicy) bool {
	if policy.NonIdempotent || !policy.Idempotent {
		return strings.TrimSpace(policy.ExceptionReason) != ""
	}
	return strings.TrimSpace(policy.Key) != ""
}

func NewResultAck(commandID, host, status, receiptHash string, timestamp time.Time) ResultAck {
	return ResultAck{
		SchemaVersion: SchemaVersion,
		Kind:          EventResultRecorded,
		CommandID:     strings.TrimSpace(commandID),
		Host:          strings.TrimSpace(host),
		Status:        strings.TrimSpace(status),
		Recorded:      true,
		ReceiptHash:   strings.TrimSpace(receiptHash),
		Timestamp:     timestamp,
	}
}

func NewEvent(kind, host, commandID, status string, timestamp time.Time, payload json.RawMessage) Event {
	return Event{
		SchemaVersion: SchemaVersion,
		Kind:          strings.TrimSpace(kind),
		Host:          strings.TrimSpace(host),
		CommandID:     strings.TrimSpace(commandID),
		Status:        strings.TrimSpace(status),
		Timestamp:     timestamp,
		Payload:       payload,
	}
}

func DefaultSubscribedKinds() []string {
	return []string{EventWorkerPoll, EventResultRecorded, EventCommandAccepted, EventControllerHeartbeat}
}

func NormalizeSubscribedKinds(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || strings.ContainsAny(value, "\r\n\t\x00") || len(value) > 96 || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) == 0 {
		return DefaultSubscribedKinds()
	}
	sort.Strings(out)
	return out
}

func SessionMatchesEvent(session CommandSession, event Event) bool {
	if session.Host != "" && event.Host != "" && !strings.EqualFold(session.Host, event.Host) {
		return false
	}
	if len(session.SubscribedKinds) == 0 {
		return true
	}
	for _, kind := range session.SubscribedKinds {
		if kind == "*" || strings.EqualFold(kind, event.Kind) {
			return true
		}
	}
	return false
}
