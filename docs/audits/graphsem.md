# Graph Semantics Module Audit

## Module

- Proposed package: `github.com/egidinas/signalforge/graphsem`
- Source path: private shared staging `go/graphsem`
- Public problem solved: Provide neutral semantic signal, graph composition, and source-catalogue contracts that downstream telemetry applications can share without importing the private shared repository.
- Public API summary: Semantic signal/series/graph types, source-family/category/kind/role/hint vocabularies, in-memory signal catalogue, source catalogue validation, source signal selection, and global catalogue summaries.

## Clean-Room Review

- Private inputs excluded: Operational `sourcecatalogue` mesh/fleet discovery, `livebus`, node identity, hostnames, gateway endpoints, SMB paths, command/provoke surfaces, leases, receipts, runtime topology, and real lab fixtures were not imported.
- Fixtures/examples included: Unit tests use synthetic `fixture_*` IDs, generic event-bus transport labels, and fictional source subjects only.
- Fixtures/examples rejected: Existing private tests containing real hostnames, serial labels, fixture names, SMB paths, and operator routes.
- Renames performed: Package path set to `github.com/egidinas/signalforge/graphsem`; package name retained.
- Compatibility aliases needed: None for the initial private consumer proof; consumers can alias the import name to `graphsem` during migration.

## Public Build

- Test command: `go test ./graphsem`
- Dependency check: Standard library only.
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: Sprint 12 planner/coordinator audit
- Date: 2026-05-12
- Decision: promote-private
- Notes: This is a bounded extraction of the semantic catalogue contract. Operational discovery loaders, mesh adapters, command authority, and private runtime routes remain downstream.
