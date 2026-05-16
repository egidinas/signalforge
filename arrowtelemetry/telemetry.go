package arrowtelemetry

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/egidinas/signalforge/contracts"
	"github.com/parquet-go/parquet-go"
)

const (
	SchemaName    = "signalforge.telemetry.arrow.v2"
	TransportMIME = "application/vnd.apache.arrow.stream"
	BatchSize     = 10000
)

var dictionaryString = &arrow.DictionaryType{IndexType: arrow.PrimitiveTypes.Uint16, ValueType: arrow.BinaryTypes.String}

var TelemetrySchema = arrow.NewSchema([]arrow.Field{
	{Name: "timestamp_ns", Type: arrow.PrimitiveTypes.Int64},
	{Name: "sensor", Type: dictionaryString},
	{Name: "value", Type: arrow.PrimitiveTypes.Float64, Nullable: true},
	{Name: "unit", Type: dictionaryString},
	{Name: "campaign_id", Type: dictionaryString},
	{Name: "source", Type: dictionaryString},
	{Name: "series_role", Type: dictionaryString},
	{Name: "signal_kind", Type: dictionaryString},
	{Name: "source_family", Type: dictionaryString},
	{Name: "quality", Type: dictionaryString},
	{Name: "state", Type: dictionaryString, Nullable: true},
}, nil)

type SignalMeta struct {
	Unit         string
	Source       string
	SeriesRole   string
	SignalKind   string
	SourceFamily string
}

type ParquetRow struct {
	TimestampNS  int64    `parquet:"timestamp_ns"`
	Sensor       string   `parquet:"sensor"`
	Value        *float64 `parquet:"value,optional"`
	Unit         string   `parquet:"unit"`
	CampaignID   string   `parquet:"campaign_id"`
	Source       string   `parquet:"source"`
	SeriesRole   string   `parquet:"series_role"`
	SignalKind   string   `parquet:"signal_kind"`
	SourceFamily string   `parquet:"source_family"`
	Quality      string   `parquet:"quality"`
	State        *string  `parquet:"state,optional"`
}

func MetadataFromGraph(model contracts.GraphModel) map[string]SignalMeta {
	meta := map[string]SignalMeta{}
	for _, lane := range model.Lanes {
		for _, series := range lane.Series {
			meta[series.ID] = SignalMeta{
				Unit:       series.Units,
				Source:     series.Source,
				SeriesRole: series.Role,
				SignalKind: "numeric",
			}
		}
	}
	if model.GraphWall != nil {
		for _, section := range model.GraphWall.Sections {
			for _, card := range section.Cards {
				for _, signal := range card.Signals {
					item := meta[signal.ID]
					if signal.Unit != "" {
						item.Unit = signal.Unit
					}
					if signal.Source != "" {
						item.Source = signal.Source
					}
					if signal.Role != "" {
						item.SeriesRole = signal.Role
					}
					if signal.Kind != "" {
						item.SignalKind = signal.Kind
					}
					if signal.SourceFamily != "" {
						item.SourceFamily = signal.SourceFamily
					}
					meta[signal.ID] = item
				}
			}
		}
	}
	return meta
}

func WriteCampaign(path, campaignID string, samples <-chan contracts.TelemetrySample, meta map[string]SignalMeta) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	f, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		return err
	}
	arrowTempPath := f.Name()
	fClosed := false
	commit := false
	defer func() {
		if !fClosed {
			_ = f.Close()
		}
		if !commit {
			_ = os.Remove(arrowTempPath)
		}
	}()

	parquetPath := strings.TrimSuffix(path, ".arrow") + ".parquet"
	parquetTemp, err := os.CreateTemp(dir, filepath.Base(parquetPath)+".*.tmp")
	if err != nil {
		return err
	}
	parquetTempPath := parquetTemp.Name()
	parquetTempClosed := false
	defer func() {
		if !parquetTempClosed {
			_ = parquetTemp.Close()
		}
		if !commit {
			_ = os.Remove(parquetTempPath)
		}
	}()

	arrowWriter := ipc.NewWriter(f, ipc.WithSchema(TelemetrySchema))
	parquetWriter := parquet.NewGenericWriter[ParquetRow](parquetTemp)
	arrowWriterClosed := false
	parquetWriterClosed := false
	defer func() {
		if !arrowWriterClosed {
			_ = arrowWriter.Close()
		}
		if !parquetWriterClosed {
			_ = parquetWriter.Close()
		}
	}()

	builder := array.NewRecordBuilder(memory.DefaultAllocator, TelemetrySchema)
	defer builder.Release()

	ts := builder.Field(0).(*array.Int64Builder)
	sensor := builder.Field(1).(*array.BinaryDictionaryBuilder)
	value := builder.Field(2).(*array.Float64Builder)
	unit := builder.Field(3).(*array.BinaryDictionaryBuilder)
	campaign := builder.Field(4).(*array.BinaryDictionaryBuilder)
	source := builder.Field(5).(*array.BinaryDictionaryBuilder)
	role := builder.Field(6).(*array.BinaryDictionaryBuilder)
	kind := builder.Field(7).(*array.BinaryDictionaryBuilder)
	sourceFamily := builder.Field(8).(*array.BinaryDictionaryBuilder)
	quality := builder.Field(9).(*array.BinaryDictionaryBuilder)
	state := builder.Field(10).(*array.BinaryDictionaryBuilder)

	var parquetRows []ParquetRow
	rowCount := 0

	flush := func() error {
		if rowCount == 0 {
			return nil
		}
		// Write Arrow Batch
		record := builder.NewRecord()
		if err := arrowWriter.Write(record); err != nil {
			record.Release()
			return err
		}
		record.Release()

		// Write Parquet Batch
		if _, err := parquetWriter.Write(parquetRows); err != nil {
			return err
		}

		// Reset for next batch
		parquetRows = parquetRows[:0]
		rowCount = 0
		return nil
	}

	for sample := range samples {
		parsed, err := time.Parse(time.RFC3339, sample.Timestamp)
		if err != nil {
			return fmt.Errorf("sample timestamp parse error: %w", err)
		}
		timestampNS := parsed.UnixNano()

		signalIDs := sortedFloatKeys(sample.Signals)
		for _, signalID := range signalIDs {
			item := meta[signalID]
			v := roundTelemetryValue(signalID, item.Unit, sample.Signals[signalID])

			ts.Append(timestampNS)
			_ = appendDict(sensor, signalID)
			value.Append(v)
			_ = appendDict(unit, item.Unit)
			_ = appendDict(campaign, campaignID)
			_ = appendDict(source, item.Source)
			_ = appendDict(role, item.SeriesRole)
			_ = appendDefault(kind, item.SignalKind, "numeric")
			_ = appendDict(sourceFamily, item.SourceFamily)
			_ = appendDict(quality, sample.Quality)
			state.AppendNull()

			parquetRows = append(parquetRows, ParquetRow{
				TimestampNS: timestampNS, Sensor: signalID, Value: ptrFloat64(v), Unit: item.Unit,
				CampaignID: campaignID, Source: item.Source, SeriesRole: item.SeriesRole,
				SignalKind: item.SignalKind, SourceFamily: item.SourceFamily, Quality: sample.Quality,
			})
			rowCount++
		}

		stateIDs := sortedStringKeys(sample.States)
		for _, signalID := range stateIDs {
			item := meta[signalID]
			ts.Append(timestampNS)
			_ = appendDict(sensor, signalID)
			value.AppendNull()
			_ = appendDefault(unit, item.Unit, "state")
			_ = appendDict(campaign, campaignID)
			_ = appendDict(source, item.Source)
			_ = appendDict(role, item.SeriesRole)
			_ = appendDefault(kind, item.SignalKind, "state")
			_ = appendDict(sourceFamily, item.SourceFamily)
			_ = appendDict(quality, sample.Quality)
			_ = appendDict(state, sample.States[signalID])

			parquetRows = append(parquetRows, ParquetRow{
				TimestampNS: timestampNS, Sensor: signalID, Unit: item.Unit,
				CampaignID: campaignID, Source: item.Source, SeriesRole: item.SeriesRole,
				SignalKind: "state", SourceFamily: item.SourceFamily, Quality: sample.Quality,
				State: ptrString(sample.States[signalID]),
			})
			rowCount++
		}

		if rowCount >= BatchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	if err := flush(); err != nil {
		return err
	}

	if err := arrowWriter.Close(); err != nil {
		return err
	}
	arrowWriterClosed = true
	if err := parquetWriter.Close(); err != nil {
		return err
	}
	parquetWriterClosed = true
	if err := f.Close(); err != nil {
		return err
	}
	fClosed = true
	if err := parquetTemp.Close(); err != nil {
		return err
	}
	parquetTempClosed = true

	gzipTempPath, err := writeGzipSidecarTemp(arrowTempPath, path+".gz")
	if err != nil {
		return err
	}
	defer func() {
		if !commit {
			_ = os.Remove(gzipTempPath)
		}
	}()

	if err := os.Rename(arrowTempPath, path); err != nil {
		return err
	}
	if err := os.Rename(parquetTempPath, parquetPath); err != nil {
		return err
	}
	if err := os.Rename(gzipTempPath, path+".gz"); err != nil {
		return err
	}
	commit = true
	return nil
}

func writeGzipSidecar(path string) error {
	tmpPath, err := writeGzipSidecarTemp(path, path+".gz")
	if err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path+".gz"); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func writeGzipSidecarTemp(path, finalPath string) (string, error) {
	src, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.CreateTemp(filepath.Dir(finalPath), filepath.Base(finalPath)+".*.tmp")
	if err != nil {
		return "", err
	}
	tmpPath := dst.Name()
	gz, err := gzip.NewWriterLevel(dst, gzip.BestCompression)
	if err != nil {
		return "", cleanupTemp(tmpPath, dst.Close())
	}
	if _, err := io.Copy(gz, src); err != nil {
		return "", cleanupTemp(tmpPath, errors.Join(err, gz.Close(), dst.Close()))
	}
	if err := gz.Close(); err != nil {
		return "", cleanupTemp(tmpPath, errors.Join(err, dst.Close()))
	}
	if err := dst.Close(); err != nil {
		return "", cleanupTemp(tmpPath, err)
	}
	return tmpPath, nil
}

func cleanupTemp(path string, err error) error {
	return errors.Join(err, os.Remove(path))
}

func roundTelemetryValue(signalID, unit string, value float64) float64 {
	if value == 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return value
	}
	id := strings.ToLower(signalID)
	unit = strings.ToLower(unit)
	switch {
	case strings.HasSuffix(id, "_count") || strings.HasSuffix(id, "_counter"):
		return math.Round(value)
	case unit == "state":
		return value
	case unit == "degc" || unit == "w" || unit == "v" || unit == "a" || unit == "db" || unit == "%" || strings.Contains(id, "_pct"):
		return roundDecimal(value, 3)
	case unit == "ms" || strings.Contains(id, "latency"):
		return roundDecimal(value, 3)
	case strings.Contains(unit, "mbar") || strings.Contains(unit, "pa"):
		return roundSignificant(value, 7)
	default:
		return roundSignificant(value, 7)
	}
}

func roundDecimal(value float64, places int) float64 {
	scale := math.Pow10(places)
	return math.Round(value*scale) / scale
}

func roundSignificant(value float64, digits int) float64 {
	if value == 0 {
		return 0
	}
	magnitude := math.Floor(math.Log10(math.Abs(value)))
	scale := math.Pow10(digits - 1 - int(magnitude))
	return math.Round(value*scale) / scale
}

func appendDefault(builder *array.BinaryDictionaryBuilder, value, fallback string) error {
	if value == "" {
		return appendDict(builder, fallback)
	}
	return appendDict(builder, value)
}

func appendDict(builder *array.BinaryDictionaryBuilder, value string) error {
	if value == "" {
		value = "unknown"
	}
	return builder.AppendString(value)
}

func ptrFloat64(value float64) *float64 {
	return &value
}

func ptrString(value string) *string {
	return &value
}

func sortedFloatKeys(values map[string]float64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
