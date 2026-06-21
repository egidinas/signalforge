# State Machine Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/statemachine`
- Source path: generic state timeline helper.
- Public problem solved: Provide deterministic observation-to-transition primitives that downstream systems can use to expose state timelines with exact event timestamps.
- Public API summary: `Observation`, `Transition`, `TransitionRule`, `Model`, `CompiledModel`, and `Tracker`.

## Clean-Room Review

- Private inputs excluded: plant-specific state names, OPC UA nodes, HMI gates, command authority, operator workflow, and capture data.
- Fixtures/examples included: Synthetic machine names and states only.
- Fixtures/examples rejected: TVac authority labels and controller traces; those remain downstream.
- Renames performed: none.
- Compatibility aliases needed: Downstream projects keep their own model projection and safety policy.

## Public Build

- Test command: `go test ./statemachine`; `go test ./...`; `git diff --check`
- Dependency check: Standard library only.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: TVac supervisor extraction
- Date: 2026-06-21
- Decision: accepted-public-core
- Notes: Package records observed state changes only. Command execution, authority classification, and private state labels stay in downstream projects. Repeated model lookups use `Model.Compile()` so downstream status loops can avoid repeated transition-slice scans.
