# Semantic Overlay Web Audit

## Module

- Proposed package: `signalforge-web` semantic overlay helpers.
- Source path: `web/src/catalogue/semanticOverlay.ts`
- Public problem solved: Keep operator aliases, fixture notes, local labels, and
  visibility preferences separate from canonical signal catalogues while still
  using stable semantic targets.
- Public API summary: `SemanticOverlayTarget`, `SemanticOverlayEntry`,
  `SemanticOverlayBundle`, `semanticOverlayTargetKey`,
  `normalizeSemanticOverlayBundle`, `overlayEntryForTarget`,
  `upsertSemanticOverlayEntry`, `removeSemanticOverlayEntry`,
  `loadSemanticOverlay`, `saveSemanticOverlay`, and `useSemanticOverlay`.

## Clean-Room Review

- Private inputs excluded: No lab topology, captures, endpoint defaults,
  credentials, private controller inventory, or fixture-specific alias data are
  included.
- Fixtures/examples included: Unit tests use synthetic targets such as
  `tec-76` only as stable public-safe examples.
- Fixtures/examples rejected: No private Loom overlay or deployment files were
  imported.
- Renames performed: The API is intentionally generic and not Meerstetter-,
  Loom-, or Gossamer-specific.
- Compatibility aliases needed: None.

## Public Build

- Test command: `npm test -- semanticOverlay`
- Build command: `npm run build`
- Local replace directives present: no
- Generated artifacts tracked: no

## Review

- Reviewer: Codex implementation pass
- Date: 2026-05-19
- Decision: promote as public web primitive
- Notes: Downstream apps may store overlays in their own namespace and merge
  aliases/notes at presentation time without mutating the canonical source
  catalogue.
