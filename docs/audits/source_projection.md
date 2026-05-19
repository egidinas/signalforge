# Source Projection Web Audit

## Module

- Proposed package: SignalForge source projection helpers.
- Source paths: `graphsem/source_projection.go` and
  `web/src/catalogue/projection.ts`
- Public problem solved: Provide one deterministic grouping/projection contract
  for all source families, so MeCom/TEC, OPC UA, CAN DBC, NI shared variables,
  Niagara, and derived signals can use the same catalogue-to-UI tree rules.
- Public API summary: Go `graphsem.SignalProjectionBundle` validation/tree
  helpers and web `SignalProjectionBundle`, `SignalProjectionMapping`,
  `SignalProjectionSignalRef`, `normalizeSignalProjectionBundle`,
  `validateSignalProjectionBundle`, `buildSignalProjectionTree`,
  `signalProjectionPathKey`, and `signalProjectionRefKey`.

## Contract

- Source-specific discovery produces source references:
  - MeCom/TEC: parameter ID, instance/channel, source device, unit, access,
    default limits, help text, and counterpart links.
  - OPC UA: NodeId, browse path, data type, access level, engineering unit,
    description, range, and source server/catalogue ID.
  - CAN DBC: DBC message/signal identity, CAN ID, multiplexer context,
    scale/offset/unit/min/max/comment metadata, and observed bus/source ID.
- Projection maps those source references into stable tree paths. It does not
  replace source truth and does not invent values.
- Source catalogue rows may carry source-agnostic grouping hints
  (`group_key`, `group_label`, `instance_key`, `sort_key`,
  `counterpart_group`, and `counterpart_trace_ids`). Projections should prefer
  those fields over family-specific string parsing.
- Primary mappings are unique. Missing primary mappings and duplicate primary
  mappings are validation errors.
- If source catalogues are supplied, trace references must resolve against the
  served source catalogue entries.
- Secondary mappings are allowed only with an explicit review reason.
- User aliases and fixture notes remain in the separate semantic overlay
  contract. They may decorate projection nodes but must not mutate the canonical
  catalogue.

## Clean-Room Review

- Private inputs excluded: No real OPC UA endpoints, DBC files, Meerstetter
  serial numbers, fixture labels, route defaults, or captured values are
  included.
- Fixtures/examples included: Unit tests use synthetic MeCom IDs, a synthetic
  OPC UA NodeId, and a synthetic DBC signal reference.
- Fixtures/examples rejected: No private Loom catalogues, real bus names, or
  lab-specific UI projections were imported.
- Renames performed: The API is intentionally source-family neutral and uses
  `source_family`, `source_id`, `trace_id`, and `signal_id` instead of
  product-specific names.
- Compatibility aliases needed: Downstream Meerstetter projection JSON can keep
  `primary_mappings`, `secondary_mappings`, `ids`, and `trace_ids`; the
  normalizer converts them into canonical `mappings` with `signal_refs`.

## Public Build

- Test command: `go test ./graphsem`; `npm test -- --runTestsByPath
  tests/projection.test.ts`
- Typecheck command: `npm run typecheck`
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: Codex implementation pass
- Date: 2026-05-19
- Decision: promote as public web primitive
- Notes: Downstream apps should consume this projection contract for grouping
  and tree rendering. Live discovery may add source catalogue entries or update
  quality/freshness, but it must not replace curated catalogue truth.
