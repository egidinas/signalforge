package graphsem

import (
	"strings"
	"testing"
)

func TestSignalDictionaryBundleValidatesSupersetFixture(t *testing.T) {
	bundle := signalDictionaryFixture()

	if err := ValidateSignalDictionaryBundle(bundle); err != nil {
		t.Fatalf("signal dictionary should validate: %v", err)
	}
}

func TestSignalDictionaryBundleRejectsUnknownRefs(t *testing.T) {
	bundle := signalDictionaryFixture()
	bundle.GraphWall[0].TraceID = "fixture.can.decoded.missing"

	err := ValidateSignalDictionaryBundle(bundle)
	if err == nil || !strings.Contains(err.Error(), "unknown source trace") {
		t.Fatalf("expected unknown trace rejection, got %v", err)
	}
}

func TestSignalDictionaryBundleRejectsActiveRoutesWithoutLease(t *testing.T) {
	bundle := signalDictionaryFixture()
	bundle.Routes[0].LeaseRequired = false
	bundle.Routes[0].AccessMode = "lease_required"

	err := ValidateSignalDictionaryBundle(bundle)
	if err == nil || !strings.Contains(err.Error(), "active access requires lease_required") {
		t.Fatalf("expected active route lease rejection, got %v", err)
	}
}

func TestSignalDictionaryBundleRejectsBrokenRouteLinks(t *testing.T) {
	bundle := signalDictionaryFixture()
	bundle.Routes[0].RingID = "missing-ring"

	err := ValidateSignalDictionaryBundle(bundle)
	if err == nil || !strings.Contains(err.Error(), "unknown ring_id") {
		t.Fatalf("expected broken ring link rejection, got %v", err)
	}
}

func TestSignalDictionaryBundleRejectsMuxWithoutMuxMetadata(t *testing.T) {
	bundle := signalDictionaryFixture()
	bundle.Routes[0].DBC.MuxSwitch = ""

	err := ValidateSignalDictionaryBundle(bundle)
	if err == nil || !strings.Contains(err.Error(), "mux_switch") {
		t.Fatalf("expected mux metadata rejection, got %v", err)
	}
}

func TestSignalDictionaryBundleValidatesDefinitionProfiles(t *testing.T) {
	bundle := signalDictionaryFixture()
	lddFamily := SourceFamily("fixture_laser_driver")
	bundle.DefinitionProfiles = append(bundle.DefinitionProfiles,
		SignalDefinitionProfile{
			ID:             "fixture.ldd130x.v1",
			DisplayName:    "Fixture LDD-130x",
			System:         "fixture_protocol",
			Family:         "laser_driver",
			SubFamily:      "ldd",
			Variant:        "ldd_130x",
			Version:        "v1",
			SourceFamilies: []SourceFamily{lddFamily},
			Classification: SignalDictionaryPublicSynthetic,
		},
		SignalDefinitionProfile{
			ID:             "fixture.ldd1321.v1",
			DisplayName:    "Fixture LDD-1321",
			System:         "fixture_protocol",
			Family:         "laser_driver",
			SubFamily:      "ldd",
			Variant:        "ldd_1321",
			Version:        "v1",
			SourceFamilies: []SourceFamily{lddFamily},
		},
	)
	bundle.Catalogues = append(bundle.Catalogues, SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "fixture_ldd",
		SourceFamily:  lddFamily,
		DisplayName:   "Fixture laser driver",
		DefinitionRef: "fixture.ldd130x.v1",
		Entries: []SourceCatalogueRow{
			{
				TraceID:        "fixture.ldd.channel_01.status",
				DisplayName:    "Fixture LDD status",
				ValueType:      "int",
				ValueTable:     map[string]string{"0": "off", "1": "ready"},
				Encoding:       &SignalValueEncoding{Kind: "enum", DataType: "uint16"},
				Access:         "subscribe",
				Category:       CategoryStatus,
				Kind:           KindState,
				Role:           RoleMonitor,
				DefinitionRef:  "fixture.ldd130x.v1",
				SourceEvidence: []string{"synthetic fixture"},
			},
		},
	})
	bundle.DomainMetadata = append(bundle.DomainMetadata, SignalDomainMetadata{
		DefinitionRef: "fixture.ldd130x.v1",
		System:        "fixture_protocol",
		Family:        "laser_driver",
		SubFamily:     "ldd",
		Required:      []string{"definition_ref", "encoding.kind"},
	})

	if err := ValidateSignalDictionaryBundle(bundle); err != nil {
		t.Fatalf("signal dictionary with definition profiles should validate: %v", err)
	}
}

func TestSignalDictionaryBundleRejectsUnknownDefinitionRefs(t *testing.T) {
	bundle := signalDictionaryFixture()
	bundle.DefinitionProfiles = []SignalDefinitionProfile{
		{ID: "fixture.device.v1", System: "fixture_protocol", Family: "fixture_device"},
	}
	bundle.Catalogues[0].DefinitionRef = "fixture.missing.v1"

	err := ValidateSignalDictionaryBundle(bundle)
	if err == nil || !strings.Contains(err.Error(), "unknown definition_ref") {
		t.Fatalf("expected unknown catalogue definition_ref rejection, got %v", err)
	}

	bundle = signalDictionaryFixture()
	bundle.DefinitionProfiles = []SignalDefinitionProfile{
		{ID: "fixture.device.v1", System: "fixture_protocol", Family: "fixture_device"},
	}
	bundle.Catalogues[0].Entries[0].DefinitionRef = "fixture.missing.v1"

	err = ValidateSignalDictionaryBundle(bundle)
	if err == nil || !strings.Contains(err.Error(), "unknown definition_ref") {
		t.Fatalf("expected unknown row definition_ref rejection, got %v", err)
	}
}

func TestSignalDictionaryBundleRejectsIncompleteDefinitionProfiles(t *testing.T) {
	bundle := signalDictionaryFixture()
	bundle.DefinitionProfiles = []SignalDefinitionProfile{
		{ID: "fixture.incomplete.v1", System: "fixture_protocol"},
	}

	err := ValidateSignalDictionaryBundle(bundle)
	if err == nil || !strings.Contains(err.Error(), "family is required") {
		t.Fatalf("expected incomplete definition rejection, got %v", err)
	}
}

func signalDictionaryFixture() SignalDictionaryBundle {
	canRef := SignalProjectionRef{
		TraceID:      "fixture.can.decoded.temperature_c",
		SourceID:     "fixture_can",
		SourceFamily: SourceFamilyCanDbc,
		Role:         RoleMonitor,
		Unit:         "degC",
	}
	tmtcRef := SignalProjectionRef{
		TraceID:      "fixture.tmtc.tm.temperature_c",
		SourceID:     "fixture_tmtc",
		SourceFamily: SourceFamilyTMTC,
		Role:         RoleMonitor,
		Unit:         "degC",
	}
	return SignalDictionaryBundle{
		SchemaVersion:  CurrentSignalDictionarySchemaVersion,
		FixtureID:      "public.synthetic.superset",
		Classification: SignalDictionaryPublicSynthetic,
		DefinitionProfiles: []SignalDefinitionProfile{
			{
				ID:             "fixture.can.v1",
				DisplayName:    "Fixture CAN definition",
				System:         "fixture_protocol",
				Family:         "can",
				SubFamily:      "dbc",
				Version:        "v1",
				SourceFamilies: []SourceFamily{SourceFamilyCanDbc},
			},
			{
				ID:             "fixture.tmtc.v1",
				DisplayName:    "Fixture TMTC definition",
				System:         "fixture_protocol",
				Family:         "tmtc",
				Version:        "v1",
				SourceFamilies: []SourceFamily{SourceFamilyTMTC},
			},
		},
		Catalogues: []SourceCatalogue{
			{
				SchemaVersion: CurrentSourceCatalogueSchemaVersion,
				SourceID:      "fixture_can",
				SourceFamily:  SourceFamilyCanDbc,
				DisplayName:   "Fixture decoded CAN",
				DefinitionRef: "fixture.can.v1",
				Entries: []SourceCatalogueRow{
					{
						TraceID:     "fixture.can.decoded.temperature_c",
						DisplayName: "Fixture decoded temperature",
						Unit:        "degC",
						ValueType:   "float",
						Access:      "subscribe",
						Category:    CategoryThermal,
						Kind:        KindContinuous,
						Role:        RoleMonitor,
						TargetID:    "fixture.can.temperature.target",
						Encoding:    &SignalValueEncoding{Kind: "scalar", DataType: "float32", Scale: 0.1, RawUnit: "degC"},
					},
				},
			},
			{
				SchemaVersion: CurrentSourceCatalogueSchemaVersion,
				SourceID:      "fixture_tmtc",
				SourceFamily:  SourceFamilyTMTC,
				DisplayName:   "Fixture TMTC",
				DefinitionRef: "fixture.tmtc.v1",
				Entries: []SourceCatalogueRow{
					{
						TraceID:     "fixture.tmtc.tm.temperature_c",
						DisplayName: "Fixture telemetry temperature",
						Unit:        "degC",
						ValueType:   "float",
						Access:      "subscribe",
						Category:    CategoryThermal,
						Kind:        KindContinuous,
						Role:        RoleMonitor,
					},
				},
			},
		},
		Projections: []SignalProjectionBundle{
			{
				SchemaVersion: CurrentSourceProjectionSchemaVersion,
				Namespace:     "fixture",
				Mappings: []SignalProjectionMapping{
					{
						ID:   "fixture-thermal",
						Kind: ProjectionPrimary,
						Path: []SignalProjectionPathSegment{
							{ID: "thermal", Label: "Thermal", Order: 10},
							{ID: "fixture", Label: "Fixture"},
						},
						SignalRefs: []SignalProjectionRef{canRef, tmtcRef},
					},
				},
			},
		},
		SemanticGroups: []SignalSemanticGroup{
			{ID: "thermal", Label: "Thermal", SourceFamilies: []SourceFamily{SourceFamilyCanDbc, SourceFamilyTMTC}, DefaultOpen: true},
			{ID: "thermal.fixture", Label: "Fixture", ParentID: "thermal", SignalRefs: []SignalProjectionRef{canRef, tmtcRef}},
		},
		Routes: []SignalRouteContract{
			{
				RouteID:       "fixture.can.mux.thermal",
				Operation:     SignalRouteMux,
				Label:         "Fixture CAN thermal mux",
				TransportKind: "can",
				AccessMode:    "lease_required",
				AuthorityMode: "dry_run_fixture",
				LeaseRequired: true,
				OperatorAck:   true,
				InputRefs:     []SignalProjectionRef{canRef},
				OutputRefs:    []SignalProjectionRef{tmtcRef},
				RingID:        "fixture.can.ring",
				DecimationID:  "fixture.envelope",
				DBC: &SignalDBCMetadata{
					BusID:       "fixture-bus-a",
					MessageName: "FixtureThermalStatus",
					FrameID:     "0x120",
					DLC:         8,
					MuxSwitch:   "mode",
					MuxCase:     "thermal",
					SignalName:  "temperature_c",
					ValueTable: map[string]string{
						"0": "standby",
						"1": "thermal",
					},
					Layout: &SignalDBCLayout{
						StartBit:  16,
						BitLength: 16,
						ByteOrder: "little_endian",
						Factor:    0.1,
						RawUnit:   "degC",
					},
				},
			},
		},
		Rings: []SignalRingProfile{
			{
				ID:              "fixture.can.ring",
				Label:           "Fixture CAN ring",
				SourceID:        "fixture_can",
				SignalRefs:      []SignalProjectionRef{canRef},
				CapacitySamples: 4096,
				RetentionPolicy: "drop_oldest",
				SequenceField:   "sequence",
				WatermarkField:  "watermark",
				DroppedField:    "dropped_count",
				FreshnessMS:     500,
			},
		},
		Decimations: []SignalDecimationProfile{
			{
				ID:                  "fixture.envelope",
				Algorithm:           "min_max_lttb_state_preserving",
				AppliesToKinds:      []SignalKind{KindContinuous, KindState, KindDiscrete},
				MaxPoints:           600,
				MinMaxEnvelope:      true,
				EventPreserving:     true,
				StateSpanPreserving: true,
			},
		},
		GraphWall: []SignalGraphWallTarget{
			{
				TargetID:     "fixture.graph.thermal",
				Label:        "Fixture thermal lane",
				Lane:         "thermal",
				Role:         "primary",
				SourceID:     "fixture_can",
				SourceFamily: SourceFamilyCanDbc,
				TraceID:      "fixture.can.decoded.temperature_c",
				ProjectionID: "fixture-thermal",
				RouteID:      "fixture.can.mux.thermal",
				RingID:       "fixture.can.ring",
				DecimationID: "fixture.envelope",
				SignalRefs:   []SignalProjectionRef{canRef, tmtcRef},
				Metadata:     map[string]any{"fixture_only": true},
			},
		},
		DomainMetadata: []SignalDomainMetadata{
			{
				Domain:        SourceFamilyCanDbc,
				SourceID:      "fixture_can",
				DefinitionRef: "fixture.can.v1",
				System:        "fixture_protocol",
				Family:        "can",
				SubFamily:     "dbc",
				Required:      []string{"bus_id", "frame_id", "dbc_message", "mux_switch", "mux_case", "ring_profile"},
			},
			{
				Domain:        SourceFamilyTMTC,
				SourceID:      "fixture_tmtc",
				DefinitionRef: "fixture.tmtc.v1",
				System:        "fixture_protocol",
				Family:        "tmtc",
				Required:      []string{"packet_kind", "sequence", "correlation_id", "authority_mode"},
			},
		},
	}
}
