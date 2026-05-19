package graphsem

import "time"

type SignalID string
type SeriesID string
type SectionID string
type EventID string
type SegmentID string
type GraphID string
type DUTID string
type CycleID string
type PhaseID string

type SourceFamily string

const (
	SourceFamilyCanDbc           SourceFamily = "can_dbc"
	SourceFamilyTec              SourceFamily = "tec"
	SourceFamilyMeComTec         SourceFamily = "mecom_tec"
	SourceFamilyThermalChamber   SourceFamily = "thermal_chamber"
	SourceFamilyOPCUA            SourceFamily = "opcua"
	SourceFamilyNISharedVariable SourceFamily = "ni_shared_variable"
	SourceFamilyOtherDiscovered  SourceFamily = "other_discovered"
	SourceFamilyBeckhoffTrace    SourceFamily = "beckhoff_trace"
	SourceFamilyVacuumPressure   SourceFamily = "vacuum_pressure"
	SourceFamilyNiagara          SourceFamily = "niagara"
	SourceFamilyExtraTemperature SourceFamily = "extra_temperature"
	SourceFamilyFunctionalTest   SourceFamily = "functional_test"
	SourceFamilyDigitalIO        SourceFamily = "digital_io"
	SourceFamilyRelaySwitching   SourceFamily = "relay_switching"
	SourceFamilySequencer        SourceFamily = "sequencer"
	SourceFamilyTMTC             SourceFamily = "tmtc"
	SourceFamilyDerived          SourceFamily = "derived"
	SourceFamilyHistoricalCsv    SourceFamily = "historical_csv"
	SourceFamilyHistoricalTdms   SourceFamily = "historical_tdms"
	SourceFamilyHistoricalHdf5   SourceFamily = "historical_hdf5"
	SourceFamilyDataLens         SourceFamily = "data_lens"
)

type Availability string

const (
	AvailabilityFixed      Availability = "fixed"
	AvailabilityOptional   Availability = "optional"
	AvailabilityDiscovered Availability = "discovered"
	AvailabilityDerived    Availability = "derived"
)

type SignalCategory string

const (
	CategoryThermal        SignalCategory = "thermal"
	CategoryElectrical     SignalCategory = "electrical"
	CategoryPower          SignalCategory = "power"
	CategoryMotion         SignalCategory = "motion"
	CategoryAlignment      SignalCategory = "alignment"
	CategoryStatus         SignalCategory = "status"
	CategoryFault          SignalCategory = "fault"
	CategoryCounter        SignalCategory = "counter"
	CategoryControl        SignalCategory = "control"
	CategoryCommunications SignalCategory = "communications"
	CategoryFunctionalTest SignalCategory = "functional_test"
	CategorySequence       SignalCategory = "sequence"
	CategoryTiming         SignalCategory = "timing"
	CategoryRaw            SignalCategory = "raw"
	CategoryUncategorized  SignalCategory = "uncategorized"
	CategoryOther          SignalCategory = "other"
)

type SignalKind string

const (
	KindContinuous SignalKind = "continuous"
	KindDiscrete   SignalKind = "discrete"
	KindState      SignalKind = "state"
	KindBoolean    SignalKind = "boolean"
	KindCounter    SignalKind = "counter"
	KindEnum       SignalKind = "enum"
	KindUnknown    SignalKind = "unknown"
)

type GraphHint string

const (
	HintLine  GraphHint = "line"
	HintStep  GraphHint = "step"
	HintState GraphHint = "state"
	HintPoint GraphHint = "point"
	HintBand  GraphHint = "band"
)

type SeriesRole string

const (
	RoleActual     SeriesRole = "actual"
	RoleExpected   SeriesRole = "expected"
	RoleGhost      SeriesRole = "ghost"
	RoleTarget     SeriesRole = "target"
	RoleReference  SeriesRole = "reference"
	RoleStatus     SeriesRole = "status"
	RoleAnnotation SeriesRole = "annotation"
)

type GraphMode string

const (
	GraphModeLiveHero           GraphMode = "live_hero"
	GraphModeHistoricalHeroLike GraphMode = "historical_hero_like"
	GraphModeHistoricalFlexible GraphMode = "historical_flexible"
	GraphModeChamberOverview    GraphMode = "chamber_overview"
)

type ExpectationProvenance string

const (
	ExpectationNA           ExpectationProvenance = ""
	ExpectationIdealPlan    ExpectationProvenance = "ideal_plan"
	ExpectationCommanded    ExpectationProvenance = "commanded"
	ExpectationRecalculated ExpectationProvenance = "recalculated"
	ExpectationInferredHist ExpectationProvenance = "inferred_history"
)

type VisibilityLevel string

const (
	VisibilityGlobal VisibilityLevel = "global"
	VisibilityFocus  VisibilityLevel = "focus"
	VisibilityDetail VisibilityLevel = "detail"
)

type AxisMode string

const (
	AxisModeWallClock AxisMode = "wall_clock"
)

type DUTPresentationMode string

const (
	DUTPresentationAll      DUTPresentationMode = "all"
	DUTPresentationSelected DUTPresentationMode = "selected"
)

type SectionKind string

const (
	SectionKindHero            SectionKind = "hero"
	SectionKindSupport         SectionKind = "support"
	SectionKindDetail          SectionKind = "detail"
	SectionKindStatusSwimlane  SectionKind = "status_swimlane"
	SectionKindFaultSwimlane   SectionKind = "fault_swimlane"
	SectionKindChamberOverview SectionKind = "chamber_overview"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
	SeverityFatal    Severity = "fatal"
)

type ConfidenceLevel string

const (
	ConfidenceObserved ConfidenceLevel = "observed"
	ConfidenceDerived  ConfidenceLevel = "derived"
	ConfidenceInferred ConfidenceLevel = "inferred"
)

type EmphasisLevel string

const (
	EmphasisPrimary   EmphasisLevel = "primary"
	EmphasisSecondary EmphasisLevel = "secondary"
	EmphasisSubdued   EmphasisLevel = "subdued"
	EmphasisAlarm     EmphasisLevel = "alarm"
)

type EventDisplayMode string

const (
	EventDisplayPoint        EventDisplayMode = "point"
	EventDisplayBand         EventDisplayMode = "band"
	EventDisplayPointAndBand EventDisplayMode = "point_and_band"
)

type SignalRole string

const (
	RoleMonitor SignalRole = "monitor"
	RoleControl SignalRole = "control"
)

type SemanticSignal struct {
	SignalID            SignalID
	CanonicalName       string
	RawName             string
	DisplayName         string
	Subsystem           string
	Category            SignalCategory
	Kind                SignalKind
	Unit                string
	Role                SignalRole
	GroupKey            string
	GroupLabel          string
	InstanceKey         string
	SortKey             string
	CounterpartGroup    string
	CounterpartTraceIDs []string
	SourceFamily        SourceFamily
	Availability        Availability
	SourceInstance      string
	DUTID               DUTID
	DefaultHint         GraphHint
}

type SeriesSpec struct {
	SeriesID         SeriesID
	Signal           SemanticSignal
	Role             SeriesRole
	Hint             GraphHint
	Expectation      ExpectationProvenance
	SectionID        SectionID
	VisibleByDefault bool
	Emphasis         EmphasisLevel
	Condensed        bool
	OutlierEligible  bool
	ValueTable       map[string]string `json:"value_table,omitempty"`
	ValueColors      map[string]string `json:"value_colors,omitempty"`
}

type PhaseContext struct {
	PhaseID     string `json:"phase_id"`
	DisplayName string `json:"display_name"`
	IsCycleEnd  bool   `json:"is_cycle_end"`
}

type EventMarker struct {
	EventID      EventID
	Kind         string
	DisplayName  string
	StartTime    time.Time
	EndTime      *time.Time
	Severity     Severity
	Importance   int
	Visibility   VisibilityLevel
	DUTID        DUTID
	CycleID      CycleID
	PhaseID      PhaseID
	DisplayMode  EventDisplayMode
	SourceFamily SourceFamily
	Confidence   ConfidenceLevel
	Phase        *PhaseContext `json:"phase,omitempty"`
	Metadata     map[string]string
}

type PlannedSegment struct {
	SegmentID  SegmentID
	Kind       string
	Label      string
	StartTime  time.Time
	EndTime    time.Time
	DUTID      DUTID
	CycleID    CycleID
	PhaseID    PhaseID
	Provenance ExpectationProvenance
	Metadata   map[string]string
}

type TimeWindow struct {
	Start time.Time
	End   time.Time
}

type TimeContext struct {
	AxisMode        AxisMode
	GlobalWindow    TimeWindow
	FocusWindow     *TimeWindow
	AbsoluteNow     *time.Time
	Elapsed         time.Duration
	PredictedFinish *time.Time
	ScheduleDrift   *time.Duration
}

type DUTPresentation struct {
	Mode             DUTPresentationMode
	SelectedDUTs     []DUTID
	HighlightedDUTs  []DUTID
	ShowAllByDefault bool
	ShowMedianTrace  bool
	ShowGroupContext bool
}

type GraphSection struct {
	SectionID    SectionID
	Title        string
	Kind         SectionKind
	DefaultOrder int
	Hideable     bool
	Reorderable  bool
	Collapsible  bool
}

type DataRef struct {
	RefKind string
	RefID   string
}

const (
	RefKindVictoriaMetrics = "victoriametrics"
	RefKindLiveStream      = "live_stream"
	RefKindHistoricalQuery = "historical_query"
	RefKindAggregate       = "aggregate"
)

type GraphSeries struct {
	Spec SeriesSpec
	Data DataRef
}

type HeroGraphModel struct {
	GraphID         GraphID
	Title           string
	Mode            GraphMode
	Time            TimeContext
	Sections        []GraphSection
	DUTPresentation DUTPresentation
	Series          []GraphSeries
	Events          []EventMarker
	Planned         []PlannedSegment
	Metadata        map[string]string
}
