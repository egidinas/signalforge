// Package statemachine provides small, deterministic state timeline
// primitives for telemetry systems that need to record observed state changes.
//
// The package is intentionally application-neutral: it does not command
// hardware, own safety policy, or infer private process semantics. Downstream
// projects provide their own state names, transition rules, and authority
// decisions.
package statemachine
