// Package ddd contains small tactical DDD primitives shared across Podzone services.
//
// The package is intentionally not a framework. Service domains still own their
// aggregate-specific repositories, commands, queries, and application use cases.
// Use these types only for cross-cutting concepts that must stay consistent:
// aggregate identity/version, event recording, domain errors, clocks, and ID
// generation.
//
// Typical aggregate usage:
//
//	base, err := ddd.NewAggregateBase(orderID, 0)
//	if err != nil {
//		return nil, err
//	}
//	order := &Order{AggregateBase: base}
//
// Prefer explicit domain behavior methods over setters. Repositories may use
// SetAggregateVersion after rehydrating a saved aggregate, but repositories
// should not emit domain events while loading snapshots/documents.
package ddd
