package graphwall

import (
	"fmt"
	"strings"
)

type TileKind string

const (
	TileTrend   TileKind = "trend"
	TileState   TileKind = "state"
	TileCommand TileKind = "command"
	TileLog     TileKind = "log"
)

const (
	AggregateTemperature = "temperature"
	AggregateTarget      = "target"
	AggregatePower       = "power"
	AggregateState       = "state"
	AggregateEvents      = "events"
	AggregateOther       = "other"
)

const (
	AxisTemperatureC = "temperature_c"
	AxisPowerW       = "power_w"
	AxisUnitLinear   = "unit_linear"
	AxisStateLane    = "state_lane"
	AxisEvent        = "event"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// TileConfig is the graph-wall persistence contract. It references discovered
// targets by ID so assignments can be resolved after each fresh discovery pass.
type TileConfig struct {
	WallID   string         `json:"wall_id"`
	TileID   string         `json:"tile_id"`
	Kind     TileKind       `json:"kind"`
	TargetID string         `json:"target_id"`
	Position Position       `json:"position"`
	Options  map[string]any `json:"options,omitempty"`
}

type Assignment[T any] struct {
	WallID   string         `json:"wall_id"`
	TileID   string         `json:"tile_id"`
	Kind     TileKind       `json:"kind"`
	Target   T              `json:"target"`
	Position Position       `json:"position"`
	Options  map[string]any `json:"options,omitempty"`
}

func ResolveAssignments[T any](configs []TileConfig, targets []T, targetID func(T) string) ([]Assignment[T], error) {
	byID := make(map[string]T, len(targets))
	for _, target := range targets {
		byID[targetID(target)] = target
	}
	out := make([]Assignment[T], 0, len(configs))
	for _, tile := range configs {
		target, ok := byID[tile.TargetID]
		if !ok {
			return nil, fmt.Errorf("graphwall: tile %q references unknown target %q", tile.TileID, tile.TargetID)
		}
		out = append(out, Assignment[T]{
			WallID:   tile.WallID,
			TileID:   tile.TileID,
			Kind:     tile.Kind,
			Target:   target,
			Position: tile.Position,
			Options:  tile.Options,
		})
	}
	return out, nil
}

type TilePolicy struct {
	DefaultPoints               int      `json:"default_points"`
	MaxPoints                   int      `json:"max_points"`
	LiveTileMinRefreshMS        int      `json:"live_tile_min_refresh_ms"`
	HistoryTileMaxCount         int      `json:"history_tile_max_count"`
	ViewportPrefetchPX          int      `json:"viewport_prefetch_px"`
	TileBufferMaxEntries        int      `json:"tile_buffer_max_entries"`
	TileBufferTTLMS             int      `json:"tile_buffer_ttl_ms"`
	ResolutionLevels            []string `json:"resolution_levels"`
	SubscriberRole              string   `json:"subscriber_role"`
	SharedTimebaseRequired      bool     `json:"shared_timebase_required"`
	LegendMayAffectPlotWidth    bool     `json:"legend_may_affect_plot_width"`
	MalformedSVGPathHardFailure bool     `json:"malformed_svg_path_hard_failure"`
}

func DenseOperatorTilePolicy() TilePolicy {
	return TilePolicy{
		DefaultPoints:               900,
		MaxPoints:                   3600,
		LiveTileMinRefreshMS:        1000,
		HistoryTileMaxCount:         96,
		ViewportPrefetchPX:          640,
		TileBufferMaxEntries:        192,
		TileBufferTTLMS:             90000,
		ResolutionLevels:            []string{"raw", "1m", "5m", "15m"},
		SubscriberRole:              "operator_supervisor",
		SharedTimebaseRequired:      true,
		LegendMayAffectPlotWidth:    false,
		MalformedSVGPathHardFailure: true,
	}
}

type Interaction struct {
	SharedTimeline       bool `json:"shared_timeline"`
	SharedCrosshair      bool `json:"shared_crosshair"`
	VerticalGridAligned  bool `json:"vertical_grid_aligned"`
	SingleTimeAxis       bool `json:"single_time_axis"`
	CursorInspect        bool `json:"cursor_inspect"`
	CrosshairAllCards    bool `json:"crosshair_all_cards"`
	LegendDoesNotResize  bool `json:"legend_does_not_resize"`
	TouchGraphFocusMode  bool `json:"touch_graph_focus_mode"`
	TreeMayCollapseAside bool `json:"tree_may_collapse_aside"`
}

func DenseOperatorInteraction() Interaction {
	return Interaction{
		SharedTimeline:       true,
		SharedCrosshair:      true,
		VerticalGridAligned:  true,
		SingleTimeAxis:       true,
		CursorInspect:        true,
		CrosshairAllCards:    true,
		LegendDoesNotResize:  true,
		TouchGraphFocusMode:  true,
		TreeMayCollapseAside: true,
	}
}

type LayoutContract struct {
	Mode              string `json:"mode"`
	SameWidthCards    bool   `json:"same_width_cards"`
	VerticalScroll    bool   `json:"vertical_scroll"`
	FixedLabelRail    bool   `json:"fixed_label_rail"`
	ResponsiveFullRow bool   `json:"responsive_full_row"`
	NoNestedCards     bool   `json:"no_nested_cards"`
}

func DenseOperatorLayout() LayoutContract {
	return LayoutContract{
		Mode:              "dense_operator_wall",
		SameWidthCards:    true,
		VerticalScroll:    true,
		FixedLabelRail:    true,
		ResponsiveFullRow: true,
		NoNestedCards:     true,
	}
}

type SemanticInput struct {
	ID       string
	Name     string
	Unit     string
	Kind     string
	Metadata map[string]string
}

func SemanticAggregate(input SemanticInput) string {
	if group := metadataFirst(input.Metadata, "graph_group", "aggregate"); group != "" {
		return strings.ToLower(group)
	}
	text := strings.ToLower(strings.Join([]string{input.ID, input.Name, input.Kind, input.Unit}, " "))
	if containsAny(text, "event", "error", "fault", "warning", "status", "state", "enum", "boolean", "bool") {
		return AggregateState
	}
	if containsAny(text, "target", "setpoint", "set point", "command") {
		return AggregateTarget
	}
	if containsAny(text, "power", "current", "voltage", "output") || unitIsAny(input.Unit, "w", "kw", "a", "ma", "v", "mv") {
		return AggregatePower
	}
	if containsAny(text, "temperature", "temp") || unitIsAny(input.Unit, "degc", "c", "celsius", "k") {
		return AggregateTemperature
	}
	return AggregateOther
}

func AxisPolicyForAggregate(group string, unit string) string {
	switch strings.ToLower(strings.TrimSpace(group)) {
	case AggregateTemperature:
		return AxisTemperatureC
	case AggregatePower:
		if unit != "" {
			return AxisUnitLinear
		}
		return AxisPowerW
	case AggregateState:
		return AxisStateLane
	case AggregateEvents:
		return AxisEvent
	default:
		return AxisUnitLinear
	}
}

func CanonicalAxisUnit(unit string) string {
	switch CanonicalUnitKey(unit) {
	case "degc":
		return "temperature_c"
	case "w", "kw":
		return "power_w"
	case "v", "mv":
		return "voltage_v"
	case "a", "ma":
		return "current_a"
	case "percent":
		return "percent"
	default:
		return CanonicalUnitKey(unit)
	}
}

func CanonicalUnitKey(unit string) string {
	normalized := strings.ToLower(strings.TrimSpace(unit))
	normalized = strings.NewReplacer(" ", "", "_", "", "-", "", ".", "").Replace(normalized)
	switch normalized {
	case "", "1", "none":
		return ""
	case "c", "degc", "degreesc", "celsius", "°c":
		return "degc"
	case "v", "volt", "volts":
		return "v"
	case "mv", "millivolt", "millivolts":
		return "mv"
	case "w", "watt", "watts":
		return "w"
	case "kw", "kilowatt", "kilowatts":
		return "kw"
	case "mw", "milliwatt", "milliwatts":
		return "mw"
	case "a", "amp", "amps", "ampere", "amperes":
		return "a"
	case "ma", "milliamp", "milliamps", "milliampere", "milliamperes":
		return "ma"
	case "dbm":
		return "dbm"
	case "db":
		return "db"
	case "hz", "hertz":
		return "hz"
	case "%", "percent", "pct":
		return "percent"
	case "ratio", "ber":
		return "ratio"
	case "count", "counts":
		return "count"
	default:
		return normalized
	}
}

func metadataFirst(metadata map[string]string, keys ...string) string {
	for _, key := range keys {
		if strings.TrimSpace(metadata[key]) != "" {
			return strings.TrimSpace(metadata[key])
		}
	}
	return ""
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func unitIsAny(unit string, values ...string) bool {
	unit = strings.ToLower(strings.TrimSpace(unit))
	for _, value := range values {
		if unit == value {
			return true
		}
	}
	return false
}

type CleanInterval struct {
	Name        string `json:"name"`
	StepSeconds int64  `json:"step_seconds"`
}

type TimeViewportRequest struct {
	DataStartUnixNano     int64 `json:"data_start_unix_nano"`
	DataEndUnixNano       int64 `json:"data_end_unix_nano"`
	SelectedStartUnixNano int64 `json:"selected_start_unix_nano"`
	PreferredStepSeconds  int64 `json:"preferred_step_seconds,omitempty"`
}

type TimeViewport struct {
	DataStartUnixNano int64         `json:"data_start_unix_nano"`
	DataEndUnixNano   int64         `json:"data_end_unix_nano"`
	ViewStartUnixNano int64         `json:"view_start_unix_nano"`
	ViewEndUnixNano   int64         `json:"view_end_unix_nano"`
	Interval          CleanInterval `json:"interval"`
}

var CleanIntervalLadder = []CleanInterval{
	{"1 min", 60},
	{"5 min", 300},
	{"10 min", 600},
	{"15 min", 900},
	{"30 min", 1800},
	{"45 min", 2700},
	{"60 min", 3600},
	{"2 h", 7200},
	{"3 h", 10800},
	{"6 h", 21600},
	{"12 h", 43200},
	{"1 d", 86400},
	{"2 d", 172800},
	{"7 d", 604800},
	{"14 d", 1209600},
	{"30 d", 2592000},
	{"90 d", 7776000},
	{"180 d", 15552000},
	{"365 d", 31536000},
}

func CalculateViewport(req TimeViewportRequest) TimeViewport {
	rawDurationNano := req.DataEndUnixNano - req.SelectedStartUnixNano
	if rawDurationNano < 0 {
		rawDurationNano = 0
	}
	rawDurationSec := rawDurationNano / 1e9

	// Target 4 to 8 intervals
	var bestInterval CleanInterval = CleanIntervalLadder[0]
	found := false
	for _, interval := range CleanIntervalLadder {
		intervals := float64(rawDurationSec) / float64(interval.StepSeconds)
		if intervals >= 4 && intervals <= 6 {
			bestInterval = interval
			found = true
			break
		}
	}

	if !found {
		if rawDurationSec/CleanIntervalLadder[0].StepSeconds > 8 {
			// Too many intervals for smallest step, try to find the one that gives > 8 but is closest
			for i := len(CleanIntervalLadder) - 1; i >= 0; i-- {
				if float64(rawDurationSec)/float64(CleanIntervalLadder[i].StepSeconds) >= 1 {
					bestInterval = CleanIntervalLadder[i]
					break
				}
			}
		}
	}

	stepNano := bestInterval.StepSeconds * 1e9
	viewStart := (req.SelectedStartUnixNano / stepNano) * stepNano
	viewEnd := ((req.DataEndUnixNano + stepNano - 1) / stepNano) * stepNano

	return TimeViewport{
		DataStartUnixNano: req.DataStartUnixNano,
		DataEndUnixNano:   req.DataEndUnixNano,
		ViewStartUnixNano: viewStart,
		ViewEndUnixNano:   viewEnd,
		Interval:          bestInterval,
	}
}
