package graphsem

import (
	"context"
	"time"
)

type SignalFilter struct {
	DUTIDs         []DUTID
	SourceFamilies []SourceFamily
	Categories     []SignalCategory
	CanonicalLike  string
}

type EventQuery struct {
	Window   TimeWindow
	DUTIDs   []DUTID
	CycleIDs []CycleID
	PhaseIDs []PhaseID
}

type PlannedQuery struct {
	Window   TimeWindow
	DUTIDs   []DUTID
	CycleIDs []CycleID
	PhaseIDs []PhaseID
}

type DataQuery struct {
	Window         TimeWindow
	TargetPoints   int
	IncludeRaw     bool
	ResolutionHint string
}

type ModeDecisionInput struct {
	IsLiveSession            bool
	HasExpectedMetadata      bool
	HasGhostCapability       bool
	HasPlannedSegments       bool
	HasSequencerContext      bool
	HasFunctionalTestMarkers bool
	IsMultiTestSelection     bool
	IsMixedHistoricalSet     bool
}

type ModeDecider interface {
	DecideMode(ctx context.Context, in ModeDecisionInput) (GraphMode, error)
}

type SignalCatalog interface {
	ListSignals(ctx context.Context, filter SignalFilter) ([]SemanticSignal, error)
	GetSignalByID(ctx context.Context, id SignalID) (SemanticSignal, error)
	GetSignalByCanonicalName(ctx context.Context, canonicalName string, dutID DUTID) (SemanticSignal, error)
}

type LiveSignalProvider interface {
	GetActiveSignals() []string
}

type EventProvider interface {
	ListEvents(ctx context.Context, q EventQuery) ([]EventMarker, error)
}

type PlannedSegmentProvider interface {
	ListPlannedSegments(ctx context.Context, q PlannedQuery) ([]PlannedSegment, error)
}

type SeriesDataResolver interface {
	ResolveSeriesData(ctx context.Context, spec SeriesSpec, q DataQuery) (DataRef, error)
}

type HeroComposeRequest struct {
	GraphID          GraphID
	Title            string
	Window           TimeWindow
	FocusWindow      *TimeWindow
	Now              *time.Time
	SelectedDUTs     []DUTID
	HighlightedDUTs  []DUTID
	ShowAllByDefault bool
	ShowMedianTrace  bool
	ShowGroupContext bool
	Scenario         string
	Mode             GraphMode
}

type HistoricalComposeRequest struct {
	GraphID      GraphID
	Title        string
	Window       TimeWindow
	FocusWindow  *TimeWindow
	Mode         GraphMode
	SelectedDUTs []DUTID
}

type HeroComposer interface {
	ComposeHero(ctx context.Context, req HeroComposeRequest) (*HeroGraphModel, error)
}

type HistoryComposer interface {
	ComposeHistoricalHeroLike(ctx context.Context, req HistoricalComposeRequest) (*HeroGraphModel, error)
	ComposeHistoricalFlexible(ctx context.Context, req HistoricalComposeRequest) (*HeroGraphModel, error)
}
