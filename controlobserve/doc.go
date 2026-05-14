// Package controlobserve provides passive control-loop observation primitives.
//
// It deliberately does not write device parameters, trigger self tuning, or
// own operator approval policy. Downstream systems can use its observations and
// recommendations as evidence, then apply their own command authority rules.
package controlobserve
