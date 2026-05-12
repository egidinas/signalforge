package graphsem

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type MemorySignalCatalog struct {
	mu      sync.RWMutex
	signals map[SignalID]SemanticSignal
}

func NewMemorySignalCatalog() *MemorySignalCatalog {
	return &MemorySignalCatalog{
		signals: make(map[SignalID]SemanticSignal),
	}
}

func (c *MemorySignalCatalog) Register(sig SemanticSignal) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.signals[sig.SignalID] = sig
}

func (c *MemorySignalCatalog) ListSignals(ctx context.Context, filter SignalFilter) ([]SemanticSignal, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []SemanticSignal
	for _, sig := range c.signals {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if matchesSignalFilter(sig, filter) {
			results = append(results, sig)
		}
	}
	return results, nil
}

func (c *MemorySignalCatalog) GetSignalByID(ctx context.Context, id SignalID) (SemanticSignal, error) {
	if err := ctx.Err(); err != nil {
		return SemanticSignal{}, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	sig, ok := c.signals[id]
	if !ok {
		return SemanticSignal{}, fmt.Errorf("signal not found: %s", id)
	}
	return sig, nil
}

func (c *MemorySignalCatalog) GetSignalByCanonicalName(ctx context.Context, canonicalName string, dutID DUTID) (SemanticSignal, error) {
	if err := ctx.Err(); err != nil {
		return SemanticSignal{}, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, sig := range c.signals {
		if sig.CanonicalName == canonicalName && sig.DUTID == dutID {
			return sig, nil
		}
	}
	return SemanticSignal{}, fmt.Errorf("signal not found by canonical name: %s for DUT %s", canonicalName, dutID)
}

func matchesSignalFilter(sig SemanticSignal, filter SignalFilter) bool {
	if len(filter.DUTIDs) > 0 && !containsDUTID(filter.DUTIDs, sig.DUTID) {
		return false
	}
	if len(filter.Categories) > 0 && !containsSignalCategory(filter.Categories, sig.Category) {
		return false
	}
	if len(filter.SourceFamilies) > 0 && !containsSourceFamily(filter.SourceFamilies, sig.SourceFamily) {
		return false
	}
	if filter.CanonicalLike != "" && !strings.Contains(strings.ToLower(sig.CanonicalName), strings.ToLower(filter.CanonicalLike)) {
		return false
	}
	return true
}

func containsDUTID(values []DUTID, want DUTID) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsSignalCategory(values []SignalCategory, want SignalCategory) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsSourceFamily(values []SourceFamily, want SourceFamily) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
