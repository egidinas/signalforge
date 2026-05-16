package graphsem

import "testing"

func TestSourceCatalogueValidateAndResolveSelection(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		SourceFamily:  SourceFamilyMeComTec,
		DisplayName:   "Fixture source",
		Entries: []SourceCatalogueRow{
			{
				TraceID:       "fixture.channel_01.temperature_c",
				RawName:       "TEMP_CH1",
				DisplayName:   "Channel 01 temperature",
				Unit:          "degC",
				ValueType:     "float",
				Access:        "subscribe",
				GraphType:     "line",
				Category:      CategoryThermal,
				Kind:          KindContinuous,
				Role:          RoleMonitor,
				DefaultHint:   HintLine,
				SourceSubject: "telemetry.fixture.temperature",
			},
			{
				TraceID:     "fixture.channel_01.state",
				RawName:     "STATE_CH1",
				DisplayName: "Channel 01 state",
				ValueType:   "string",
				Access:      "subscribe",
				GraphType:   "step",
				Category:    CategoryStatus,
				Kind:        KindState,
				Role:        RoleMonitor,
				DefaultHint: HintStep,
			},
		},
		Capabilities: SourceCapabilities{
			SupportsLive:      true,
			MaxSignals:        2,
			RecommendedRateHz: 1,
			TransportPaths: []TransportPath{
				{PathID: "fixture_bus", PathKind: "stream", NetworkTransport: "event_bus", State: "available"},
			},
		},
	}

	if err := catalogue.Validate(); err != nil {
		t.Fatalf("catalogue did not validate: %v", err)
	}

	selection := SourceSignalSelection{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		SourceFamily:  SourceFamilyMeComTec,
		Signals: []SelectedSignal{
			{
				SignalID:      "fixture.temp.01",
				TraceID:       "fixture.channel_01.temperature_c",
				CanonicalName: "fixture.temperature.channel_01",
				DUTID:         "fixture_dut_01",
			},
		},
	}
	resolved, err := ResolveSourceSignalSelection(catalogue, selection)
	if err != nil {
		t.Fatalf("selection did not resolve: %v", err)
	}
	if len(resolved.Signals) != 1 || resolved.UnselectedCount != 1 {
		t.Fatalf("resolved selection = %#v", resolved)
	}
	signal := resolved.Signals[0]
	if signal.CanonicalName != "fixture.temperature.channel_01" {
		t.Fatalf("CanonicalName = %q", signal.CanonicalName)
	}
	if signal.SourceFamily != SourceFamilyMeComTec || signal.Category != CategoryThermal || signal.DefaultHint != HintLine {
		t.Fatalf("semantic fields = %#v", signal)
	}
}

func TestSourceCatalogueRejectsDuplicateSelectedTraceID(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		SourceFamily:  SourceFamilyMeComTec,
		Entries: []SourceCatalogueRow{
			{
				TraceID:     "fixture.channel_01.temperature_c",
				RawName:     "TEMP_CH1",
				DisplayName: "Channel 01 temperature",
				Unit:        "degC",
				ValueType:   "float",
				Access:      "subscribe",
				GraphType:   "line",
				Category:    CategoryThermal,
				Kind:        KindContinuous,
				Role:        RoleMonitor,
				DefaultHint: HintLine,
			},
		},
	}
	selection := SourceSignalSelection{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		Signals: []SelectedSignal{
			{SignalID: "fixture.temp.primary", TraceID: "fixture.channel_01.temperature_c"},
			{SignalID: "fixture.temp.alias", TraceID: "fixture.channel_01.temperature_c"},
		},
	}

	if err := selection.ValidateAgainst(catalogue); err == nil {
		t.Fatalf("expected duplicate trace_id selection to be rejected")
	}
}

func TestSourceCatalogueRejectsInvalidRows(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		SourceFamily:  SourceFamilyCanDbc,
		Entries: []SourceCatalogueRow{
			{TraceID: "fixture.frame.raw"},
			{TraceID: "fixture.frame.raw", Access: "subscribe"},
		},
	}

	if err := catalogue.Validate(); err == nil {
		t.Fatalf("expected invalid catalogue")
	}
}

func TestGlobalSourceCatalogueSummary(t *testing.T) {
	global := GlobalSourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		Catalogues: []SourceCatalogue{
			{
				SchemaVersion: CurrentSourceCatalogueSchemaVersion,
				SourceID:      "source_a",
				SourceFamily:  SourceFamilyCanDbc,
				Entries: []SourceCatalogueRow{
					{TraceID: "source_a.frame_001", Access: "subscribe"},
					{TraceID: "source_a.frame_002", Access: "subscribe"},
				},
			},
		},
		Selections: []SourceSignalSelection{
			{
				SchemaVersion: CurrentSourceCatalogueSchemaVersion,
				SourceID:      "source_a",
				Signals: []SelectedSignal{
					{SignalID: "source_a.frame_001", TraceID: "source_a.frame_001"},
				},
			},
		},
		DiscoveredCatalogues: []DiscoveredCatalogueRecord{
			{CatalogueID: "source_a_catalogue", LinkedSourceID: "source_a", DisplayName: "Source A", SourceFamily: SourceFamilyCanDbc, Status: "available"},
		},
	}

	summary, err := global.Summary()
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}
	if summary.SourceCount != 1 || summary.EntryCount != 2 || summary.SelectedCount != 1 || summary.UnselectedCount != 1 {
		t.Fatalf("summary = %#v", summary)
	}
	if summary.ByFamily[SourceFamilyCanDbc] != 2 {
		t.Fatalf("ByFamily = %#v", summary.ByFamily)
	}
}
