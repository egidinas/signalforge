# Publication Readiness

This note records the current public-release posture for SignalForge.

## Current State

- Module path: `github.com/egidinas/signalforge`
- Remote: `git@github.com:egidinas/signalforge.git`
- License: **MIT** — decided 2026-05-14.
- Tag: **v0.0.1-dev** — cut 2026-05-14 as consumer-proof baseline. Not a stable release.
- All 16 packages pass `go test ./...`. No `replace` directives in `go.mod`.

## Pending Decisions

- **Module path identity** — `github.com/egidinas/signalforge` is the working
  path. The decision on whether to move to a name-linked path (e.g.
  `github.com/jrmeyer/signalforge`) is deferred until after the Orbitworks
  technical panel. A rename before the first stable tag is low-cost.
- Approve the exact public module path before cutting a stable `v0.1.0` tag.

## Audit Note

- `docs/audits/cantrace.md` and `docs/audits/stability.md` name
  `mynaric_telemetry` as extraction source paths. These are provenance records,
  not runtime dependencies. Review before first public promotion whether the
  employer name in audit docs is acceptable or should be anonymised to a generic
  label (e.g. "prior private telemetry system").

## Public Consumer Proof

Before calling SignalForge publication-ready, the first public consumer proof should satisfy all of the following:

- A fresh clone can fetch SignalForge from the approved VCS target and module path.
- There is no local sibling `replace` directive in the published module graph.
- The safepath consumer import has moved to the approved module path.
- Focused `/data/` route positive tests pass.
- Path traversal rejection tests pass.

## Do Not Publish

- The proof-only absolute `replace` from Sprint 3 must remain out of committed public history.
- This document is advisory only and does not assert that publication has happened.
