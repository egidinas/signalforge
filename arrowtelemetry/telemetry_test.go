package arrowtelemetry

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/egidinas/signalforge/contracts"
	"github.com/parquet-go/parquet-go"
)

func TestWriteCampaignProducesCanonicalArrowRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "telemetry.arrow")
	samples := []contracts.TelemetrySample{{
		Timestamp: "2026-01-15T10:00:00Z",
		Signals: map[string]float64{
			"chamber_air_deg_c": 21.123456789,
			"tm_packet_counter": 1234.567,
			"tvac_pressure_pa":  0.000000123456789,
		},
		States:  map[string]string{"facility_interlock_state": "nominal"},
		Quality: "fresh",
	}}
	meta := map[string]SignalMeta{
		"chamber_air_deg_c":        {Unit: "degC", Source: "facility_thermal", SeriesRole: "actual", SignalKind: "measurement", SourceFamily: "facility"},
		"tm_packet_counter":        {Unit: "count", Source: "flight_computer", SeriesRole: "counter", SignalKind: "counter", SourceFamily: "packet"},
		"tvac_pressure_pa":         {Unit: "Pa", Source: "facility_vacuum", SeriesRole: "actual", SignalKind: "measurement", SourceFamily: "vacuum"},
		"facility_interlock_state": {Unit: "state", Source: "demo_quality", SeriesRole: "interlock", SignalKind: "state", SourceFamily: "quality"},
	}

	ch := make(chan contracts.TelemetrySample, len(samples))
	for _, s := range samples {
		ch <- s
	}
	close(ch)
	if err := WriteCampaign(path, "thermal_acceptance_fat", ch, meta); err != nil {
		t.Fatalf("write campaign arrow: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open arrow: %v", err)
	}
	defer f.Close()
	reader, err := ipc.NewReader(f)
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}
	defer reader.Release()
	if !reader.Next() {
		t.Fatal("expected one record")
	}
	rec := reader.Record()
	if rec.NumRows() != 4 {
		t.Fatalf("rows = %d, want 4", rec.NumRows())
	}
	ts := rec.Column(0).(*array.Int64)
	wantNS := mustParse("2026-01-15T10:00:00Z").UnixNano()
	if got := ts.Value(0); got != wantNS {
		t.Fatalf("timestamp_ns = %d, want %d", got, wantNS)
	}
	sensor := rec.Column(1).(*array.Dictionary)
	if sensor.ValueStr(0) != "chamber_air_deg_c" || sensor.ValueStr(1) != "tm_packet_counter" || sensor.ValueStr(2) != "tvac_pressure_pa" || sensor.ValueStr(3) != "facility_interlock_state" {
		t.Fatalf("sensor column = %q, %q, %q, %q", sensor.ValueStr(0), sensor.ValueStr(1), sensor.ValueStr(2), sensor.ValueStr(3))
	}
	values := rec.Column(2).(*array.Float64)
	if got := values.Value(0); got != 21.123 {
		t.Fatalf("rounded temperature = %.12f, want 21.123", got)
	}
	if got := values.Value(1); got != 1235 {
		t.Fatalf("rounded counter = %.12f, want 1235", got)
	}
	if got := values.Value(2); got != 0.0000001234568 {
		t.Fatalf("rounded pressure = %.16f, want 0.0000001234568", got)
	}
	state := rec.Column(10).(*array.Dictionary)
	if !state.IsNull(0) || state.ValueStr(3) != "nominal" {
		t.Fatalf("state column null/value mismatch")
	}
	assertParquetNullParity(t, strings.TrimSuffix(path, ".arrow")+".parquet")
}

func TestWriteCampaignPreservesExistingArtifactsOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "telemetry.arrow")
	parquetPath := filepath.Join(dir, "telemetry.parquet")
	gzipPath := path + ".gz"
	arrowSentinel := []byte("old arrow")
	parquetSentinel := []byte("old parquet")
	gzipSentinel := []byte("old gzip")
	if err := os.WriteFile(path, arrowSentinel, 0o644); err != nil {
		t.Fatalf("write sentinel arrow: %v", err)
	}
	if err := os.WriteFile(parquetPath, parquetSentinel, 0o644); err != nil {
		t.Fatalf("write sentinel parquet: %v", err)
	}
	if err := os.WriteFile(gzipPath, gzipSentinel, 0o644); err != nil {
		t.Fatalf("write sentinel gzip: %v", err)
	}

	ch := make(chan contracts.TelemetrySample, 2)
	ch <- contracts.TelemetrySample{
		Timestamp: "2026-01-15T10:00:00Z",
		Signals:   map[string]float64{"temperature": 21.5},
		Quality:   "fresh",
	}
	ch <- contracts.TelemetrySample{
		Timestamp: "not-a-timestamp",
		Signals:   map[string]float64{"temperature": 22.0},
		Quality:   "fresh",
	}
	close(ch)
	if err := WriteCampaign(path, "campaign", ch, map[string]SignalMeta{"temperature": {Unit: "degC"}}); err == nil {
		t.Fatal("WriteCampaign succeeded with malformed timestamp")
	}
	assertFileBytes(t, path, arrowSentinel)
	assertFileBytes(t, parquetPath, parquetSentinel)
	assertFileBytes(t, gzipPath, gzipSentinel)
}

func mustParse(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s = %q, want %q", path, string(got), string(want))
	}
}

func assertParquetNullParity(t *testing.T, path string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open parquet: %v", err)
	}
	defer f.Close()
	reader := parquet.NewGenericReader[ParquetRow](f)
	defer reader.Close()

	rows := make([]ParquetRow, 4)
	n, err := reader.Read(rows)
	if err != nil && err != io.EOF {
		t.Fatalf("read parquet: %v", err)
	}
	if n != 4 {
		t.Fatalf("parquet rows = %d, want 4", n)
	}
	for i := 0; i < 3; i++ {
		if rows[i].Value == nil {
			t.Fatalf("parquet numeric row %d has nil value", i)
		}
		if rows[i].State != nil {
			t.Fatalf("parquet numeric row %d has state %q, want null", i, *rows[i].State)
		}
	}
	if rows[3].Value != nil {
		t.Fatalf("parquet state row value = %v, want null", *rows[3].Value)
	}
	if rows[3].State == nil || *rows[3].State != "nominal" {
		t.Fatalf("parquet state row state = %v, want nominal", rows[3].State)
	}
}
