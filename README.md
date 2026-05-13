# SignalForge

SignalForge is a Go toolkit for building resilient telemetry and telecommand
applications with source discovery, tile-based graphing, command ownership,
archive/replay, simulation, and transport adapters for mixed hardware and
multi-node systems.

This repository starts as a clean public baseline. It must not import history,
fixtures, deployment defaults, screenshots, private procedures, captures, DBCs,
hostnames, IP routes, credentials, or customer/project identifiers from any
internal staging repository.

## Initial Scope

SignalForge should grow from small, audited neutral primitives:

- safe file/path helpers, starting with [`signalforge/safepath`](safepath)
  for lexical containment checks; callers that dereference paths through
  untrusted directories still need symlink-aware handling;
- JSON and JSONL file helpers, starting with [`signalforge/jsonfile`](jsonfile)
  for pretty-printed JSON writes and append-only line records;
- graph wall policy helpers, starting with [`signalforge/graphwall`](graphwall)
  for deterministic tile assignment, axis normalization, layout defaults, and
  viewport selection;
- numeric clamping helpers, starting with [`signalforge/mathutil`](mathutil)
  for ordered values and duration bounds;
- semantic graph and source-catalogue contracts, starting with
  [`signalforge/graphsem`](graphsem) for neutral signal vocabularies,
  catalogue validation, and selection resolution;
- reusable control-program contracts, starting with
  [`signalforge/controlprogram`](controlprogram) for target fanout, cyclic
  setpoint programs, hold durations, and validation without binding to a device
  vendor or fieldbus protocol;
- compact agent/context encoding;
- source catalogue and discovery-tree contracts;
- graph/tile query contracts;
- command session and event/audit log primitives;
- archive/replay interfaces;
- transport adapter interfaces;
- fictional fixture and simulation helpers.

Hardware-specific behavior, live deployment configuration, real lab topology,
and private operator workflows belong in downstream applications.

## Import Rule

Every copied module needs a completed audit note before it enters this repo.
Use [docs/module_audit_template.md](docs/module_audit_template.md). The default
answer for a questionable module is to leave it out until it can be expressed as
public, general-purpose software.

## Public Safety

Before every public tag, run the checklist in
[docs/public_safety_checklist.md](docs/public_safety_checklist.md). A fresh
clone must build and test using public dependencies only, with no local sibling
repositories or private configuration.

For the current publication-status note and pending decisions, see
[docs/publication_readiness.md](docs/publication_readiness.md).
