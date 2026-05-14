# Ring Buffer Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/ringbuf`
- Source path: `meerstetter-go/canring` and `private-telemetry/internal/utils/circular_buffer.go`
- Public problem solved: Provide generic in-memory and optionally file-backed bounded rings for edge telemetry bootstrap, fallback, and late-owner replay.
- Public API summary: Generic RAM `Buffer[T]` with caller-provided byte sizer,
  deterministic oldest-first eviction, `Push`, `Snapshot`, `Drain`, `Len`,
  `Bytes`, and `MaxBytes`.

## Clean-Room Review

- Private inputs excluded: PiXtend service paths, private bus names, TEC parameter semantics, captures, and deployment defaults.
- Fixtures/examples included: Synthetic records with generic sequence and timestamp fields.
- Fixtures/examples rejected: Real CAN, MeCom, TMTC, or lab-node records.
- Renames performed: Telemetry-row-specific circular buffer becomes a generic
  byte-budgeted RAM ring.
- Compatibility aliases needed: Meerstetter-Go and Loom keep protocol-specific
  ring readers, file-backed persistence, sequence deduplication, and bootstrap
  replay around the generic core.

## Public Build

- Test command: `go test ./ringbuf`; `go test ./...`; `git diff --check`
- Dependency check: Standard library only.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: migration backlog
- Date: 2026-05-14
- Decision: accepted-public-core
- Notes: This first extraction is RAM-only and deterministic. File-backed rings,
  device-ring merge/dedup, monotonic sequence contracts, and power-loss replay
  remain downstream adapter work.
