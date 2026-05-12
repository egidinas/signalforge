# Publication Readiness

This note records the current public-release posture for SignalForge.

## Current State

- Private staging module path: `github.com/egidinas/signalforge`
- Private staging remote: `git@github.com:egidinas/signalforge.git`
- No public remote has been selected yet.
- No publication tag has been cut yet.
- No license has been selected yet.

## Pending Decisions

- Approve the exact public module path before the first public tag.
- The private staging module path is not a public-release commitment; approve
  the exact public module path before the first public tag.
- Choose the license before the first public push or tag.

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
