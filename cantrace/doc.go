// Package cantrace contains driver-neutral helpers for classic CAN frame
// normalization, byte parsing, DLC handling, and monitor formatting.
//
// The package intentionally does not own vendor handles, SocketCAN/Kvaser
// lifecycle, route ownership, DBC parsing, or private capture defaults. Those
// remain in downstream adapters.
package cantrace
