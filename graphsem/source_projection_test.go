package graphsem

import (
	"strings"
	"testing"
)

func TestSignalProjectionValidatesAgainstSourceCatalogues(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "opcua-main",
		SourceFamily:  SourceFamilyOPCUA,
		Entries: []SourceCatalogueRow{
			{TraceID: "ns=2;s=Line1.Temp", Access: "subscribe"},
			{TraceID: "ns=2;s=Line1.TempSetpoint", Access: "read_write"},
		},
	}
	bundle := SignalProjectionBundle{
		SchemaVersion: CurrentSourceProjectionSchemaVersion,
		Namespace:     "fixture",
		Mappings: []SignalProjectionMapping{
			{
				ID:   "line1-temp",
				Kind: ProjectionPrimary,
				Path: []SignalProjectionPathSegment{
					{ID: "thermal", Label: "Thermal", Order: 10},
					{ID: "line1", Label: "Line 1"},
				},
				SignalRefs: []SignalProjectionRef{
					{TraceID: "ns=2;s=Line1.Temp", SourceID: "opcua-main", SourceFamily: SourceFamilyOPCUA},
					{TraceID: "ns=2;s=Line1.TempSetpoint", SourceID: "opcua-main", SourceFamily: SourceFamilyOPCUA},
				},
			},
		},
	}

	if err := ValidateSignalProjectionBundle(bundle, SignalProjectionValidationOptions{Catalogues: []SourceCatalogue{catalogue}}); err != nil {
		t.Fatalf("projection should validate: %v", err)
	}
	tree := BuildSignalProjectionTree(bundle)
	if len(tree) != 1 || tree[0].ID != "thermal" || len(tree[0].Children) != 1 {
		t.Fatalf("unexpected tree: %#v", tree)
	}
}

func TestSignalProjectionRejectsDuplicatePrimaryAndUnknownTrace(t *testing.T) {
	catalogue := SourceCatalogue{
		SchemaVersion: CurrentSourceCatalogueSchemaVersion,
		SourceID:      "thermal-can",
		SourceFamily:  SourceFamilyCanDbc,
		Entries:       []SourceCatalogueRow{{TraceID: "can_dbc:0x123:Status.object_temperature", Access: "subscribe"}},
	}
	bundle := SignalProjectionBundle{
		SchemaVersion: CurrentSourceProjectionSchemaVersion,
		Mappings: []SignalProjectionMapping{
			{
				ID:   "primary-a",
				Kind: ProjectionPrimary,
				Path: []SignalProjectionPathSegment{{ID: "thermal", Label: "Thermal"}},
				SignalRefs: []SignalProjectionRef{
					{TraceID: "can_dbc:0x123:Status.object_temperature", SourceID: "thermal-can", SourceFamily: SourceFamilyCanDbc},
				},
			},
			{
				ID:   "primary-b",
				Kind: ProjectionPrimary,
				Path: []SignalProjectionPathSegment{{ID: "thermal-copy", Label: "Thermal Copy"}},
				SignalRefs: []SignalProjectionRef{
					{TraceID: "can_dbc:0x123:Status.object_temperature", SourceID: "thermal-can", SourceFamily: SourceFamilyCanDbc},
				},
			},
			{
				ID:   "unknown",
				Kind: ProjectionPreview,
				Path: []SignalProjectionPathSegment{{ID: "unknown", Label: "Unknown"}},
				SignalRefs: []SignalProjectionRef{
					{TraceID: "can_dbc:0x123:Status.missing", SourceID: "thermal-can", SourceFamily: SourceFamilyCanDbc},
				},
			},
		},
	}

	err := ValidateSignalProjectionBundle(bundle, SignalProjectionValidationOptions{Catalogues: []SourceCatalogue{catalogue}})
	if err == nil {
		t.Fatal("expected projection validation to fail")
	}
	if !strings.Contains(err.Error(), "duplicates primary projection") && !strings.Contains(err.Error(), "unknown source trace") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignalProjectionRequiresReasonForSecondaryMappings(t *testing.T) {
	bundle := SignalProjectionBundle{
		SchemaVersion: CurrentSourceProjectionSchemaVersion,
		Mappings: []SignalProjectionMapping{
			{
				ID:         "duplicate-view",
				Kind:       ProjectionSecondary,
				Path:       []SignalProjectionPathSegment{{ID: "diagnostics", Label: "Diagnostics"}},
				SignalRefs: []SignalProjectionRef{{SignalID: "tec.76.ch1.temperature"}},
			},
		},
	}

	err := ValidateSignalProjectionBundle(bundle, SignalProjectionValidationOptions{})
	if err == nil || !strings.Contains(err.Error(), "review reason") {
		t.Fatalf("expected secondary reason failure, got %v", err)
	}
}
