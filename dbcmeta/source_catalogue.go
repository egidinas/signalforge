package dbcmeta

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/egidinas/signalforge/graphsem"
)

// SourceCatalogueOptions controls how parsed DBC metadata is exposed through
// the shared graphsem source-catalogue contract.
type SourceCatalogueOptions struct {
	SourceID       string
	DisplayName    string
	SourceSubject  string
	HistoryPath    string
	TargetID       string
	TransportPaths []graphsem.TransportPath
}

// DBCSignalTraceID is the stable source-local identity for one DBC signal.
func DBCSignalTraceID(message Message, signal Signal) string {
	return fmt.Sprintf("can_dbc:0x%X:%s.%s", message.ID, message.Name, signal.Name)
}

// SourceCatalogueFromFile converts parsed DBC metadata into the shared
// graphsem source catalogue used by UI projections and graph assignment.
func SourceCatalogueFromFile(db *File, options SourceCatalogueOptions) (graphsem.SourceCatalogue, error) {
	if db == nil {
		return graphsem.SourceCatalogue{}, fmt.Errorf("dbc file is nil")
	}
	sourceID := strings.TrimSpace(options.SourceID)
	if sourceID == "" {
		sourceID = "dbc"
	}
	displayName := strings.TrimSpace(options.DisplayName)
	if displayName == "" {
		displayName = "DBC CAN catalogue"
	}

	catalogue := graphsem.SourceCatalogue{
		SchemaVersion: graphsem.CurrentSourceCatalogueSchemaVersion,
		SourceID:      sourceID,
		SourceFamily:  graphsem.SourceFamilyCanDbc,
		DisplayName:   displayName,
		Entries:       dbcSourceRows(db, options),
		Capabilities: graphsem.SourceCapabilities{
			SupportsLive:         true,
			SupportsHistory:      strings.TrimSpace(options.HistoryPath) != "",
			SupportsMetadataOnly: true,
			TransportPaths:       append([]graphsem.TransportPath(nil), options.TransportPaths...),
		},
	}
	if err := catalogue.Validate(); err != nil {
		return graphsem.SourceCatalogue{}, err
	}
	return catalogue, nil
}

func dbcSourceRows(db *File, options SourceCatalogueOptions) []graphsem.SourceCatalogueRow {
	ids := make([]uint32, 0, len(db.Messages))
	for id := range db.Messages {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	var rows []graphsem.SourceCatalogueRow
	for _, id := range ids {
		message := db.Messages[id]
		if message == nil {
			continue
		}
		for _, signal := range message.Signals {
			rows = append(rows, graphsem.SourceCatalogueRow{
				TraceID:       DBCSignalTraceID(*message, signal),
				RawName:       message.Name + "." + signal.Name,
				DisplayName:   humanDBCLabel(message.Name, signal.Name),
				Unit:          signal.Unit,
				ValueType:     dbcSignalValueType(signal),
				Access:        "read",
				GraphSource:   "can_dbc",
				GraphType:     string(dbcGraphHint(signal)),
				Category:      dbcSignalCategory(signal),
				Kind:          dbcSignalKind(signal),
				Role:          graphsem.RoleMonitor,
				GroupKey:      dbcMessageGroupKey(*message),
				GroupLabel:    humanDBCLabel(message.Name, ""),
				InstanceKey:   dbcMessageInstanceKey(*message),
				SortKey:       dbcSignalSortKey(*message, signal),
				DefaultHint:   dbcGraphHint(signal),
				SourceSubject: strings.TrimSpace(options.SourceSubject),
				HistoryPath:   strings.TrimSpace(options.HistoryPath),
				TargetID:      strings.TrimSpace(options.TargetID),
				TargetFormat:  "can_dbc_signal",
				TargetUse:     "decode",
				DiscoveryBadges: []string{
					"dbc",
					"can",
				},
				TargetMetadata: dbcTargetMetadata(*message, signal),
				Metadata:       dbcSignalMetadata(*message, signal),
			})
		}
	}
	return rows
}

func dbcMessageGroupKey(message Message) string {
	return fmt.Sprintf("can_dbc:0x%X:%s", message.ID, message.Name)
}

func dbcMessageInstanceKey(message Message) string {
	return fmt.Sprintf("0x%X", message.ID)
}

func dbcSignalSortKey(message Message, signal Signal) string {
	return fmt.Sprintf("%08X.%03d.%s", message.ID, signal.StartBit, signal.Name)
}

func dbcTargetMetadata(message Message, signal Signal) map[string]string {
	return map[string]string{
		"can_id":       strconv.FormatUint(uint64(message.ID), 10),
		"can_id_hex":   fmt.Sprintf("0x%X", message.ID),
		"message_name": message.Name,
		"signal_name":  signal.Name,
		"sender":       message.Sender,
	}
}

func dbcSignalMetadata(message Message, signal Signal) map[string]string {
	metadata := map[string]string{
		"dlc":       strconv.Itoa(message.DLC),
		"start_bit": strconv.Itoa(signal.StartBit),
		"length":    strconv.Itoa(signal.Length),
		"endian":    endianLabel(signal.BigEndian),
		"signed":    strconv.FormatBool(signal.Signed),
		"factor":    strconv.FormatFloat(signal.Factor, 'g', -1, 64),
		"offset":    strconv.FormatFloat(signal.Offset, 'g', -1, 64),
		"min":       strconv.FormatFloat(signal.Min, 'g', -1, 64),
		"max":       strconv.FormatFloat(signal.Max, 'g', -1, 64),
	}
	if len(signal.Receivers) > 0 {
		metadata["receivers"] = strings.Join(signal.Receivers, ",")
	}
	if len(signal.ValueTable) > 0 {
		metadata["value_table"] = valueTableString(signal.ValueTable)
	}
	return metadata
}

func dbcSignalValueType(signal Signal) string {
	if signal.IsBoolean {
		return "bool"
	}
	if signal.IsEnum {
		return "enum"
	}
	return "float64"
}

func dbcSignalKind(signal Signal) graphsem.SignalKind {
	if signal.IsBoolean {
		return graphsem.KindBoolean
	}
	if signal.IsEnum {
		return graphsem.KindEnum
	}
	return graphsem.KindContinuous
}

func dbcGraphHint(signal Signal) graphsem.GraphHint {
	if signal.IsBoolean || signal.IsEnum {
		return graphsem.HintStep
	}
	return graphsem.HintLine
}

func dbcSignalCategory(signal Signal) graphsem.SignalCategory {
	unit := strings.ToLower(strings.TrimSpace(signal.Unit))
	name := strings.ToLower(signal.Name)
	switch {
	case strings.Contains(unit, "deg") || strings.Contains(unit, "c") && strings.Contains(name, "temp") || strings.Contains(name, "temperature"):
		return graphsem.CategoryThermal
	case unit == "v" || strings.Contains(name, "volt"):
		return graphsem.CategoryElectrical
	case unit == "a" || strings.Contains(name, "current"):
		return graphsem.CategoryElectrical
	case unit == "w" || strings.Contains(name, "power"):
		return graphsem.CategoryPower
	case strings.Contains(name, "fault") || strings.Contains(name, "error"):
		return graphsem.CategoryFault
	case strings.Contains(name, "status") || signal.IsBoolean || signal.IsEnum:
		return graphsem.CategoryStatus
	default:
		return graphsem.CategoryRaw
	}
}

func humanDBCLabel(messageName string, signalName string) string {
	return strings.TrimSpace(strings.ReplaceAll(messageName+" "+signalName, "_", " "))
}

func endianLabel(bigEndian bool) string {
	if bigEndian {
		return "big"
	}
	return "little"
}

func valueTableString(values map[float64]string) string {
	keys := make([]float64, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Float64s(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, strconv.FormatFloat(key, 'g', -1, 64)+"="+values[key])
	}
	return strings.Join(parts, ";")
}
