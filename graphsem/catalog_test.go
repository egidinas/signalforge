package graphsem

import (
	"context"
	"testing"
)

func TestMemorySignalCatalogFiltersSignals(t *testing.T) {
	catalogue := NewMemorySignalCatalog()
	catalogue.Register(SemanticSignal{
		SignalID:      "fixture.temperature.01",
		CanonicalName: "fixture.temperature.channel_01",
		Category:      CategoryThermal,
		SourceFamily:  SourceFamilyMeComTec,
		DUTID:         "fixture_dut_01",
	})
	catalogue.Register(SemanticSignal{
		SignalID:      "fixture.voltage.01",
		CanonicalName: "fixture.voltage.channel_01",
		Category:      CategoryElectrical,
		SourceFamily:  SourceFamilyCanDbc,
		DUTID:         "fixture_dut_02",
	})

	signals, err := catalogue.ListSignals(context.Background(), SignalFilter{
		Categories:    []SignalCategory{CategoryThermal},
		CanonicalLike: "temp",
	})
	if err != nil {
		t.Fatalf("ListSignals failed: %v", err)
	}
	if len(signals) != 1 || signals[0].SignalID != "fixture.temperature.01" {
		t.Fatalf("signals = %#v", signals)
	}

	got, err := catalogue.GetSignalByCanonicalName(context.Background(), "fixture.voltage.channel_01", "fixture_dut_02")
	if err != nil {
		t.Fatalf("GetSignalByCanonicalName failed: %v", err)
	}
	if got.SignalID != "fixture.voltage.01" {
		t.Fatalf("SignalID = %q", got.SignalID)
	}
}
