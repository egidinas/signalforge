package dbcmeta

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egidinas/signalforge/graphsem"
)

func TestLoadDirBuildsCanonicalCandidates(t *testing.T) {
	dir := t.TempDir()
	writeDBC(t, dir, "CondorMk3_7.3.1.dbc", `
BO_ 28 CTC_Status: 8 NODE
 SG_ ctc_status : 0|8@1+ (1,0) [0|15] "" RX
 SG_ ctc_cpu_usage : 8|8@1+ (1,0) [0|100] "" RX
BO_ 385 AMC_Status: 8 NODE
 SG_ amc_status : 0|8@1+ (1,0) [0|15] "" RX
`)
	writeDBC(t, dir, "CondorMk3_7.3.1 (1).dbc", `
BO_ 28 CTC_Status: 8 NODE
 SG_ ctc_status : 0|8@1+ (1,0) [0|15] "" RX
 SG_ ctc_cpu_usage : 8|8@1+ (1,0) [0|100] "" RX
BO_ 385 AMC_Status: 8 NODE
 SG_ amc_status : 0|8@1+ (1,0) [0|15] "" RX
`)
	writeDBC(t, dir, "CondorMk3_7.3.2.dbc", `
BO_ 28 CTC_Status: 8 NODE
 SG_ ctc_status : 0|8@1+ (1,0) [0|3] "" RX
 SG_ ctc_cpu_usage : 8|8@1+ (1,0) [0|10] "" RX
BO_ 385 AMC_Status: 8 NODE
 SG_ amc_status : 0|8@1+ (1,0) [0|15] "" RX
`)

	catalog, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if catalog.FileCount != 3 {
		t.Fatalf("expected 3 files, got %d", catalog.FileCount)
	}
	if catalog.CanonicalCount != 2 {
		t.Fatalf("expected 2 canonical candidates, got %d", catalog.CanonicalCount)
	}
	if catalog.RawDuplicateGroupCount != 1 {
		t.Fatalf("expected 1 raw duplicate group, got %d", catalog.RawDuplicateGroupCount)
	}
	if len(catalog.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(catalog.Candidates))
	}
}

func TestRankObservedUsesSignalPlausibilityToChooseSpecificImplementation(t *testing.T) {
	dir := t.TempDir()
	writeDBC(t, dir, "CondorMk3_A.dbc", `
BO_ 28 CTC_Status: 8 NODE
 SG_ ctc_status : 0|8@1+ (1,0) [0|15] "" RX
 SG_ ctc_cpu_usage : 8|8@1+ (1,0) [0|100] "" RX
BO_ 385 AMC_Status: 8 NODE
 SG_ amc_status : 0|8@1+ (1,0) [0|15] "" RX
`)
	writeDBC(t, dir, "CondorMk3_B.dbc", `
BO_ 28 CTC_Status: 8 NODE
 SG_ ctc_status : 0|8@1+ (1,0) [0|3] "" RX
 SG_ ctc_cpu_usage : 8|8@1+ (1,0) [0|10] "" RX
BO_ 385 AMC_Status: 8 NODE
 SG_ amc_status : 0|8@1+ (1,0) [0|15] "" RX
`)

	catalog, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	results := catalog.RankObserved([]ObservedMessage{
		{ID: 28, DLC: 8, Count: 20, SampleHex: []string{"0732"}},
		{ID: 385, DLC: 8, Count: 10, SampleHex: []string{"01"}},
	}, 2)
	if len(results) < 2 {
		t.Fatalf("expected at least 2 match results, got %d", len(results))
	}
	if results[0].Representative != "CondorMk3_A" {
		t.Fatalf("expected CondorMk3_A first, got %s", results[0].Representative)
	}
	if results[0].SignalPlausibility <= results[1].SignalPlausibility {
		t.Fatalf("expected better signal plausibility for first result: %.3f <= %.3f", results[0].SignalPlausibility, results[1].SignalPlausibility)
	}
}

func TestDetailByKeyIncludesMessageFingerprints(t *testing.T) {
	dir := t.TempDir()
	writeDBC(t, dir, "CondorMk3_7.3.3.dbc", `
BO_ 28 CTC_Status: 8 NODE
 SG_ ctc_status : 0|8@1+ (1,0) [0|15] "" RX
 SG_ ctc_cpu_usage : 8|8@1+ (1,0) [0|100] "" RX
`)
	catalog, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(catalog.Candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(catalog.Candidates))
	}
	detail, ok := catalog.DetailByKey(catalog.Candidates[0].Fingerprint)
	if !ok {
		t.Fatal("DetailByKey failed")
	}
	if detail.Path == "" {
		t.Fatal("expected detail path")
	}
	if len(detail.MessageFingerprints) != 1 {
		t.Fatalf("expected 1 message fingerprint, got %d", len(detail.MessageFingerprints))
	}
	if detail.MessageFingerprints[0].Hash == "" || detail.MessageFingerprints[0].Signature == "" {
		t.Fatal("expected populated message fingerprint hash and signature")
	}
}

func TestParseBytesDecodesValueTables(t *testing.T) {
	db, err := ParseBytes([]byte(`
BO_ 100 State: 1 ECU
 SG_ mode : 0|1@1+ (1,0) [0|1] "" NODE
VAL_ 100 mode 0 "off" 1 "on";
`))
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}
	msg := db.Messages[100]
	if msg == nil || len(msg.Signals) != 1 {
		t.Fatalf("expected parsed message and signal, got %#v", msg)
	}
	if !msg.Signals[0].IsBoolean {
		t.Fatalf("expected boolean value table, got enum=%v table=%v", msg.Signals[0].IsEnum, msg.Signals[0].ValueTable)
	}
	if got := DecodeSignal([]byte{1}, &msg.Signals[0]); got != 1 {
		t.Fatalf("DecodeSignal = %v, want 1", got)
	}
}

func TestSourceCatalogueFromFileExposesDBCSignals(t *testing.T) {
	db, err := ParseBytes([]byte(`
BO_ 291 Thermal_Status: 8 ECU
 SG_ object_temperature : 0|16@1+ (0.1,-40) [-40|120] "degC" NODE
 SG_ output_enabled : 16|1@1+ (1,0) [0|1] "" NODE
VAL_ 291 output_enabled 0 "off" 1 "on";
`))
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}

	catalogue, err := SourceCatalogueFromFile(db, SourceCatalogueOptions{
		SourceID:    "thermal-can",
		DisplayName: "Thermal CAN",
		HistoryPath: "/api/history/thermal-can",
		TransportPaths: []graphsem.TransportPath{{
			PathID:            "usb-adapter",
			PathKind:          "can",
			PhysicalTransport: "usb-can",
			State:             "hot",
		}},
	})
	if err != nil {
		t.Fatalf("SourceCatalogueFromFile: %v", err)
	}

	if catalogue.SourceFamily != graphsem.SourceFamilyCanDbc {
		t.Fatalf("SourceFamily = %q, want %q", catalogue.SourceFamily, graphsem.SourceFamilyCanDbc)
	}
	if len(catalogue.Entries) != 2 {
		t.Fatalf("expected 2 source rows, got %d", len(catalogue.Entries))
	}
	first := catalogue.Entries[0]
	if first.TraceID != "can_dbc:0x123:Thermal_Status.object_temperature" {
		t.Fatalf("unexpected trace id %q", first.TraceID)
	}
	if first.GroupKey != "can_dbc:0x123:Thermal_Status" || first.GroupLabel != "Thermal Status" || first.InstanceKey != "0x123" {
		t.Fatalf("unexpected grouping: group_key=%q group_label=%q instance_key=%q", first.GroupKey, first.GroupLabel, first.InstanceKey)
	}
	if first.SortKey != "00000123.000.object_temperature" {
		t.Fatalf("unexpected sort key %q", first.SortKey)
	}
	if first.Category != graphsem.CategoryThermal || first.Kind != graphsem.KindContinuous || first.DefaultHint != graphsem.HintLine {
		t.Fatalf("unexpected semantic classification: category=%q kind=%q hint=%q", first.Category, first.Kind, first.DefaultHint)
	}
	second := catalogue.Entries[1]
	if second.Kind != graphsem.KindBoolean || second.DefaultHint != graphsem.HintStep {
		t.Fatalf("unexpected boolean classification: kind=%q hint=%q", second.Kind, second.DefaultHint)
	}
	if second.Metadata["value_table"] != "0=off;1=on" {
		t.Fatalf("unexpected value table metadata %q", second.Metadata["value_table"])
	}
}

func TestLoadDirKeepsDifferentValueTablesSeparate(t *testing.T) {
	dir := t.TempDir()
	writeDBC(t, dir, "Controller_A.dbc", `
	BO_ 100 State: 1 ECU
	 SG_ mode : 0|2@1+ (1,0) [0|3] "" NODE
	VAL_ 100 mode 0 "off" 1 "on" 2 "fault";
	`)
	writeDBC(t, dir, "Controller_B.dbc", `
	BO_ 100 State: 1 ECU
	 SG_ mode : 0|2@1+ (1,0) [0|3] "" NODE
	VAL_ 100 mode 0 "idle" 1 "active" 2 "fault";
	`)

	catalog, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if catalog.CanonicalCount != 2 {
		t.Fatalf("expected value-table variants to stay separate, got %d canonical candidates", catalog.CanonicalCount)
	}
}

func TestLoadFilesReturnsMissingFileErrors(t *testing.T) {
	dir := t.TempDir()
	validPath := writeDBC(t, dir, "Controller.dbc", `
BO_ 100 State: 1 ECU
 SG_ mode : 0|1@1+ (1,0) [0|1] "" NODE
`)
	missingPath := filepath.Join(dir, "missing.dbc")

	catalog, err := LoadFiles([]string{validPath, missingPath}, dir)
	if err == nil {
		t.Fatal("expected missing file error")
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Fatalf("error %q does not mention missing path %q", err, missingPath)
	}
	if catalog == nil || catalog.CanonicalCount != 1 {
		t.Fatalf("expected partial catalog with one valid candidate, got %#v", catalog)
	}
}

func TestLoadDirReturnsMalformedFileErrors(t *testing.T) {
	dir := t.TempDir()
	writeDBC(t, dir, "valid.dbc", `
BO_ 100 State: 1 ECU
 SG_ mode : 0|1@1+ (1,0) [0|1] "" NODE
`)
	malformedPath := writeDBC(t, dir, "malformed.dbc", `SG_ orphan : 0|1@1+ (1,0) [0|1] "" NODE`)

	catalog, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected malformed file error")
	}
	if !strings.Contains(err.Error(), malformedPath) {
		t.Fatalf("error %q does not mention malformed path %q", err, malformedPath)
	}
	if catalog == nil || catalog.CanonicalCount != 1 {
		t.Fatalf("expected partial catalog with one valid candidate, got %#v", catalog)
	}
	if !strings.Contains(err.Error(), "no messages found") {
		t.Fatalf("expected parse error context, got %v", err)
	}
}

func writeDBC(t *testing.T, dir string, name string, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}
