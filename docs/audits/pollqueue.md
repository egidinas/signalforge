# Poll Queue Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/pollqueue`
- Source path: `meerstetter-go/mecom/pollqueue.go`
- Public problem solved: Provide reusable priority round-robin polling queues where manual requests can move values to the front without bypassing normal scheduler fairness.
- Public API summary: Generic `Queue[T,V]` with caller-provided key function,
  normal `Enqueue`, one-shot `EnqueueFront`, duplicate suppression within a
  chunk, `NextChunk`, seeded `Latest` results, and `Record` updates.

## Clean-Room Review

- Private inputs excluded: MeCom parameter IDs, TEC controller addressing, device serials, and operator policies.
- Fixtures/examples included: Synthetic parameter keys and generic values only.
- Fixtures/examples rejected: Real TEC catalogues and bus captures.
- Renames performed: MeCom `Parameter` queue becomes generic item queue with a
  string key function.
- Compatibility aliases needed: Meerstetter-Go keeps parameter-specific wrappers
  and bulk-read frame sizing.

## Public Build

- Test command: `go test ./pollqueue`; `go test ./...`; `git diff --check`
- Dependency check: Standard library only.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: migration backlog
- Date: 2026-05-14
- Decision: accepted-public-core
- Notes: Covers fair round-robin, manual front-of-queue refresh, latest-value
  initialization, and per-chunk duplicate suppression. Device-specific
  parameter catalogues and congestion policy stay downstream.
