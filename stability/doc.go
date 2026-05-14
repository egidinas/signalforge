// Package stability provides reusable telemetry stability primitives.
//
// It is intentionally protocol- and application-neutral: callers provide
// timestamped signal values, criteria, windows, and group policies; the package
// returns deterministic signal and aggregate stability state. Device-specific
// catalogues, transport paths, and operator UI concerns belong in downstream
// packages.
package stability
