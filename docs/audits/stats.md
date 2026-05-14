# Streaming Statistics Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/stats`
- Source path: `meerstetter-go/mecom/reduction.go` and related reducer helpers
- Public problem solved: Provide small streaming reduction primitives for telemetry rings and graph-wall views.
- Public API summary: `Window` accepts streaming float64 values, ignores NaN,
  and snapshots `Summary{Count, Mean, Min, Max, StdDev}` using sample standard
  deviation. `Reset` clears the window for consumer-rate reduction boundaries.

## Clean-Room Review

- Private inputs excluded: TEC-specific signal names, thermal policy, controller serials, and graph-wall layout defaults.
- Fixtures/examples included: Synthetic numeric sequences only.
- Fixtures/examples rejected: Real thermal traces and private performance data.
- Renames performed: Meerstetter `ReduceRingSamples` math becomes a generic
  `Window`; Meerstetter-specific ring-frame filtering remains downstream.
- Compatibility aliases needed: Downstream adapters choose signal-specific
  reduction policy and timestamp handling.

## Public Build

- Test command: `go test ./stats`; `go test ./...`; `git diff --check`
- Dependency check: Standard library only.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: migration backlog
- Date: 2026-05-14
- Decision: accepted-public-core
- Notes: Mathematically generic only. Domain-specific SNR choices, channel
  priorities, and graph-wall decimation policies stay downstream.
