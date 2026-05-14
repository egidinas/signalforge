# CAN Trace Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/cantrace`
- Source path: `kvaser-dual-bridge/internal/netcan` and `private-telemetry/internal/canrouter`
- Public problem solved: Provide reusable CAN raw-frame, DLC, and PCAP-oriented helpers shared by vendor-specific bridges and private Loom ingest modules.
- Public API summary: Classic CAN `Frame`, data normalization, byte-string parsing, DLC fallback/resolution against caller-provided maps, frame data formatting, and caller-provided flag naming/skip-policy helpers.

## Clean-Room Review

- Private inputs excluded: Kvaser ownership details, CANlib handles, bus names, route leases, DBC paths, private captures, hostnames, and operational defaults.
- Fixtures/examples included: Synthetic CAN frames and minimal public-safe records only.
- Fixtures/examples rejected: Real bus captures and private DBC-derived names.
- Renames performed: `bus.CANFrame` copies are represented as `cantrace.Frame`; Kvaser-specific flag constants remain downstream and are supplied to generic flag helpers by callers.
- Compatibility aliases needed: Kvaser and Loom adapters should preserve their current package-level behavior.

## Public Build

- Test command: `go test ./...`
- Dependency check: Standard library only for this extraction pass.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: migration backlog
- Date: 2026-05-14
- Decision: accepted-public-core
- Notes: Extracted the duplicated driver-neutral frame, DLC, parser, formatter, and skip-policy mechanics. Deferred PCAP/TCP raw-record stream helpers because they still encode Kvaser reverse-engineering details, packet markers, and capture-flow assumptions that need a separate fixture parity pass.

## Extraction Result

- Added `cantrace.Frame` for the duplicated CAN frame shape found in private telemetry system and Kvaser bridge sources.
- Added `NormalizeData` and `NewFrame` to centralize classic CAN DLC validation.
- Added `ParseDataBytes`, `FrameData`, `DataHex`, and `FrameHex` to replace ad-hoc CLI parsing and monitor formatting.
- Added `InferFallbackDLC` and `ResolveDLC` for generic DLC fallback behavior while keeping DBC parsing downstream.
- Added `ShouldSkipFlags`, `FlagNames`, and `FormatFlagNames` so vendor adapters can keep their own flag constants without copying formatter logic.

## Deferred Work

- Raw 64-byte Kvaser record parsing and TCP sequence reassembly.
- PCAP stream reading/writing.
- Host-control and status packet parsing.
- DBC parser integration.

Those pieces should move only after public-safe synthetic fixtures cover both the `kvaser-dual-bridge/internal/netcan` and `private-telemetry/internal/canrouter` behavior.
