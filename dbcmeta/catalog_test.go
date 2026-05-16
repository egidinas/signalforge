package dbcmeta

import (
	"os"
	"path/filepath"
	"testing"
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

func writeDBC(t *testing.T, dir string, name string, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
