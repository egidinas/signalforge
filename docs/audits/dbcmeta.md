# DBC Metadata Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/dbcmeta`
- Source path: internal DBC metadata helpers.
- Public problem solved: Provide shared, public-safe DBC catalogue metadata and grouping helpers without private DBC files or deployment defaults.
- Public API summary: DBC subset parser, CAN payload signal decoder,
  canonical DBC catalogue grouping, duplicate and semantic grouping summaries,
  per-message fingerprints, observed-traffic ranking against catalogue
  candidates, and DBC-to-`graphsem.SourceCatalogue` export.

## Clean-Room Review

- Private inputs excluded: Private DBC files, bus labels, product-specific signal names not already public, host paths, and route defaults.
- Fixtures/examples included: Synthetic DBC metadata fixtures only.
- Fixtures/examples rejected: Live fleet DBC directories and private signal catalogues.
- Renames performed: Package renamed from internal DBC metadata helpers to public `dbcmeta`; exported shapes use neutral catalogue, candidate, fingerprint, and observed-message names.
- Compatibility aliases needed: Downstream consumers may need import aliases during migration.

## Public Build

- Test command: `go test ./dbcmeta`; `go test ./...`
- Dependency check: Standard library only.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: migration backlog
- Date: 2026-05-14
- Decision: accepted-public-core
- Notes: Seeded from the richer internal implementation. The public core covers
  parsing, catalogue grouping, message fingerprints, observed-traffic ranking,
  and source-catalogue export. UI evolution helpers, private DBC directories,
  and graph provenance adapters remain downstream work.

## Extraction Result

- Extracted: `signalforge/dbcmeta`
- Verified: `go test ./dbcmeta`; `go test ./...`; `git diff --check`
- Public fixtures: Synthetic DBC strings created in tests.
- Private data copied: none.

## Deferred Work

- Rewire bridge DBC metadata consumers through the public core.
- Rewire graph provenance adapters through the public core.
- Add parity tests against scrubbed fixture outputs before archiving either legacy implementation.
