# SignalForge Web

SignalForge Web is a development-stage React package for application-neutral
telemetry UI primitives:

- signal dictionary navigation and wall assignment state;
- wall creation, selection, rename, and removal helpers;
- graph-tile fetch helpers and React hooks;
- graph-tile normalization and rendered-series helpers;
- uPlot-backed tile rendering with overlays, markers, decimation, and visual
  policy helpers.

The package deliberately stops at shared primitives. Downstream applications
own their app shells, routing, authentication, command authority, live transport
adapters, device-specific catalogues, and deployment defaults.

## Package Contract

The public entrypoint is `src/index.ts`. Consumers supply adapters for signal
catalogues, channel mapping, tile fetching, live subscriptions, formatting, and
optional command writes.

`CANONICAL_TILE_RENDERER`, `emptyGraphTile`, `normalizeGraphTile`,
`renderSeriesFromGraphTile`, and the role metadata helpers are exported from
SignalForge so downstream apps can shape product-specific telemetry into
`graph_tile.v1` while sharing the same uPlot tile path.

The library build emits `dist/signalforge-web.es.js` and declaration files under
`dist/`. The demo build is separate and emits `dist-demo/`.

## Verification

Run these checks from this directory:

```sh
npm test -- --config jest.config.cjs
npm run typecheck
npm run build
npm run demo:build
```

Before a repository tag, run the repository-level release check from the repo
root:

```sh
./scripts/release_check.sh
```
