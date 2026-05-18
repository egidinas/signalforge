// Package tilehistory provides a bounded, 1 Hz history reducer for graph tiles.
//
// It accepts timestamped numeric samples, keeps a fixed-duration ring of
// second-sized buckets, and exports per-bucket aggregates that can be rendered
// as history tiles or further downsampled by callers.
package tilehistory
