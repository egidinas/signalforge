# Module Audit: web

## Scope

Development-stage React package for public, application-neutral telemetry UI
primitives:

- signal dictionary browsing and assignment state;
- wall management state and controls;
- graph-tile client, hooks, decimation, overlays, markers, and time-axis
  helpers;
- graph-tile normalization, rendered-series ordering, and canonical renderer
  constants for downstream consumers;
- uPlot-backed tile rendering;
- standalone demo used only as a render and packaging proof.

## Public-Safety Review

- No private hostnames, IP routes, credentials, captures, screenshots, or
  deployment defaults are required by the package.
- Consumers provide all device catalogues, live transport adapters, command
  write implementations, routing, authentication, and deployment behavior.
- The demo uses local fixture-style data and must remain separate from the
  library entrypoint.
- Generated output (`dist/`, `dist-demo/`, `node_modules/`, and
  `*.tsbuildinfo`) is ignored.

## Verification

- `npm test -- --config jest.config.cjs`
- `npm run typecheck`
- `npm run build`
- `npm run demo:build`
- `./scripts/release_check.sh`

## Decision

Accepted as a public SignalForge surface while it remains limited to shared UI
primitives and adapter contracts. It is not a complete application shell.
