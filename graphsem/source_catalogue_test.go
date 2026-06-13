package graphsem

import "testing"

const fixtureCatalogueFamily SourceFamily = "fixture_device"

func TestSourceCatalogueValidateAndResolveSelection(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		SourceFamily:  fixtureCatalogueFamily,
		DisplayName:   "Fixture source",
		DefinitionRef: "fixture.device.v1",
		Entries: []SourceCatalogueRow{
			{
				TraceID:             "fixture.channel_01.temperature_c",
				RawName:             "TEMP_CH1",
				DisplayName:         "Channel 01 temperature",
				Help:                "Fixture value used by catalogue tests",
				SourceEvidence:      []string{"synthetic fixture"},
				Unit:                "degC",
				ValueType:           "float",
				Encoding:            &SignalValueEncoding{Kind: "scalar", DataType: "float32", Scale: 0.1, RawUnit: "degC"},
				Access:              "subscribe",
				GraphType:           "line",
				Category:            CategoryThermal,
				Kind:                KindContinuous,
				Role:                RoleMonitor,
				GroupKey:            "fixture/channel:1",
				GroupLabel:          "Fixture channel 1",
				InstanceKey:         "channel:1",
				SortKey:             "001.temperature",
				CounterpartGroup:    "fixture.channel_01.loop",
				CounterpartTraceIDs: []string{"fixture.channel_01.target"},
				DefaultHint:         HintLine,
				SourceSubject:       "telemetry.fixture.temperature",
			},
			{
				TraceID:     "fixture.channel_01.state",
				RawName:     "STATE_CH1",
				DisplayName: "Channel 01 state",
				ValueType:   "string",
				ValueTable:  map[string]string{"0": "standby", "1": "active"},
				Encoding:    &SignalValueEncoding{Kind: "enum", DataType: "uint8"},
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
		SourceFamily:  fixtureCatalogueFamily,
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
	if signal.SourceFamily != fixtureCatalogueFamily || signal.Category != CategoryThermal || signal.DefaultHint != HintLine {
		t.Fatalf("semantic fields = %#v", signal)
	}
	if signal.GroupKey != "fixture/channel:1" || signal.GroupLabel != "Fixture channel 1" || signal.InstanceKey != "channel:1" {
		t.Fatalf("grouping fields = %#v", signal)
	}
	if signal.CounterpartGroup != "fixture.channel_01.loop" || len(signal.CounterpartTraceIDs) != 1 || signal.CounterpartTraceIDs[0] != "fixture.channel_01.target" {
		t.Fatalf("counterpart fields = %#v", signal)
	}
}

func TestSourceCatalogueRejectsDuplicateSelectedTraceID(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		SourceFamily:  fixtureCatalogueFamily,
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

func TestSourceCatalogueRejectsInvalidEncoding(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_source",
		SourceFamily:  fixtureCatalogueFamily,
		Entries: []SourceCatalogueRow{
			{TraceID: "fixture.channel_01.flags", Access: "subscribe", Encoding: &SignalValueEncoding{BitLength: 8}},
		},
	}

	if err := catalogue.Validate(); err == nil {
		t.Fatalf("expected invalid encoding to be rejected")
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
