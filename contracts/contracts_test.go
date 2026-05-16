package contracts

import (
	"encoding/json"
	"testing"
)

func TestArchiveDatasetRef_JSON(t *testing.T) {
	ref := ArchiveDatasetRef{
		ID:              "ds-1",
		Name:            "Test Dataset",
		Format:          ArchiveFormatHDF5,
		SourceKind:      ArchiveSourceHDF5,
		CreatedUnixNano: 1000,
		StartUnixNano:   500,
		EndUnixNano:     1500,
		SignalCount:     10,
		SampleCount:     10000,
		ReadOnly:        true,
		Provenance:      map[string]string{"node": "n1"},
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var ref2 ArchiveDatasetRef
	if err := json.Unmarshal(data, &ref2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ref.ID != ref2.ID || ref.Format != ref2.Format || ref.Provenance["node"] != ref2.Provenance["node"] {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", ref2, ref)
	}
}

func TestArchiveExportRequest_JSON(t *testing.T) {
	req := ArchiveExportRequest{
		DatasetID:     "ds-1",
		Format:        ArchiveFormatFort,
		StartUnixNano: 100,
		EndUnixNano:   200,
		SignalIDs:     []string{"s1", "s2"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var req2 ArchiveExportRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if req.DatasetID != req2.DatasetID || req.Format != req2.Format || len(req.SignalIDs) != len(req2.SignalIDs) {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", req2, req)
	}
}

func TestTelecommandIdempotencyKeyIncludesArguments(t *testing.T) {
	base := Telecommand{
		TargetID:  "tec-1",
		Name:      "set_temperature",
		Arguments: map[string]any{"setpoint": 21.5, "unit": "C"},
	}
	changed := Telecommand{
		TargetID:  "tec-1",
		Name:      "set_temperature",
		Arguments: map[string]any{"setpoint": 22.0, "unit": "C"},
	}

	base.EnsureIdempotencyKey()
	changed.EnsureIdempotencyKey()
	if base.IdempotencyKey == "" {
		t.Fatal("empty idempotency key")
	}
	if base.IdempotencyKey == changed.IdempotencyKey {
		t.Fatal("argument-only command changes reused the same idempotency key")
	}
}

func TestTelecommandIdempotencyKeyArgumentsAreStable(t *testing.T) {
	a := Telecommand{
		TargetID:  "tec-1",
		Name:      "set_temperature",
		Arguments: map[string]any{"setpoint": 21.5, "unit": "C"},
	}
	b := Telecommand{
		TargetID:  "tec-1",
		Name:      "set_temperature",
		Arguments: map[string]any{"unit": "C", "setpoint": 21.5},
	}

	a.EnsureIdempotencyKey()
	b.EnsureIdempotencyKey()
	if a.IdempotencyKey != b.IdempotencyKey {
		t.Fatalf("keys differ for equivalent arguments: %q != %q", a.IdempotencyKey, b.IdempotencyKey)
	}
}

func TestTelecommandIdempotencyKeyPreservesExplicitKey(t *testing.T) {
	tc := Telecommand{TargetID: "node-1", Name: "restart", IdempotencyKey: "operator-provided"}
	tc.EnsureIdempotencyKey()
	if tc.IdempotencyKey != "operator-provided" {
		t.Fatalf("idempotency key = %q", tc.IdempotencyKey)
	}
}
