# Stability Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/stability`
- Source path: internal stability helpers.
- Public problem solved: Provide reusable stability-window, threshold, and observation primitives for telemetry systems without importing private runtime code.
- Public API summary: Stability configuration, bounded rolling buffers, signal/window/group evaluation, linear-drift and target-deviation evaluators, buffer import/export, dwell-gate transitions, and config validation.

## Clean-Room Review

- Private inputs excluded: Operator procedures, route defaults, node names, private topology, live captures, lab-specific thresholds, and UI policy.
- Fixtures/examples included: To be synthetic only.
- Fixtures/examples rejected: Any real device names, private route labels, operator screenshots, or historical captures.
- Renames performed: Pending.
- Compatibility aliases needed: Downstream systems may keep private adapters around the public core.

## Public Build

- Test command: `go test ./...`
- Dependency check: Stability package uses standard library only.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: migration backlog
- Date: 2026-05-14
- Decision: accepted-public-core
- Notes: SignalForge now owns the generic stability core. Downstream wrapper rewiring remains gated on unrelated local wrapper state.

## Source Probe

- One internal stability helper contains the preferred public-core shape: monitor types, rolling buffers, group policy evaluation, and linear drift evaluation implemented with standard-library math only.
- `private-telemetry/internal/stability` contains additional behavior that must not be lost: buffer import/export helpers and dwell-gate transition evaluation.
- Current integration risk: downstream wrapper files already have unrelated local edits, so wrapper rewiring should wait until the public core has parity tests and the dirty wrapper state is reviewed.

## Extraction Decision

- Seed the SignalForge package from the common denominator, then add the dwell-gate transition behavior as an explicit public primitive.
- Prefer the local standard-library linear-regression implementation over the private telemetry system external `montanaflynn/stats` dependency unless parity testing proves a behavior gap.
- First required tests: linear drift, deviation-from-target, N-of-M group policy, rolling-buffer eviction and snapshot, buffer import/export, and dwell-gate transition states.

## Extraction Result

- Package added: `stability`
- Tests added: linear drift, deviation from target, N-of-M group policy, monitor insufficient/stable/ramp behavior, rolling-buffer eviction, buffer export/import, dwell transition, and config validation.
- Verification: `go test ./...` passed on 2026-05-14.
- Behavior correction: the extracted package does not mark a window unstable merely because a metric is nonzero below its configured limit; it reports approaching/breached only when the configured ratio threshold is reached.
