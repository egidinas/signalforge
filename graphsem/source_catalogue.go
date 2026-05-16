package graphsem

import (
	"fmt"
	"strings"
)

const CurrentSourceCatalogueSchemaVersion = 1

type SourceCatalogue struct {
	SchemaVersion int                  `json:"schema_version"`
	SourceID      string               `json:"source_id"`
	SourceFamily  SourceFamily         `json:"source_family"`
	DisplayName   string               `json:"display_name,omitempty"`
	Page          CataloguePage        `json:"page,omitempty"`
	Entries       []SourceCatalogueRow `json:"entries"`
	Capabilities  SourceCapabilities   `json:"capabilities,omitempty"`
}

type CataloguePage struct {
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	Total      int    `json:"total,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	Filter     string `json:"filter,omitempty"`
}

type SourceCapabilities struct {
	SupportsLive          bool            `json:"supports_live,omitempty"`
	SupportsHistory       bool            `json:"supports_history,omitempty"`
	SupportsMetadataOnly  bool            `json:"supports_metadata_only,omitempty"`
	MaxSignals            int             `json:"max_signals,omitempty"`
	MaxBytesPerSecond     int             `json:"max_bytes_per_second,omitempty"`
	DefaultRateHz         float64         `json:"default_rate_hz,omitempty"`
	RecommendedRateHz     float64         `json:"recommended_rate_hz,omitempty"`
	DiscoveryCadenceSec   int             `json:"discovery_cadence_sec,omitempty"`
	SubscriptionEndpoint  string          `json:"subscription_endpoint,omitempty"`
	HistoryEndpoint       string          `json:"history_endpoint,omitempty"`
	LiveSubjects          []string        `json:"live_subjects,omitempty"`
	SelectionRequired     bool            `json:"selection_required,omitempty"`
	PoliteAccessStatement string          `json:"polite_access_statement,omitempty"`
	TransportPaths        []TransportPath `json:"transport_paths,omitempty"`
}

type TransportPath struct {
	PathID            string   `json:"path_id"`
	PathKind          string   `json:"path_kind"`
	PhysicalTransport string   `json:"physical_transport,omitempty"`
	NetworkTransport  string   `json:"network_transport,omitempty"`
	Endpoint          string   `json:"endpoint,omitempty"`
	State             string   `json:"state,omitempty"`
	Workflow          string   `json:"workflow,omitempty"`
	Notes             []string `json:"notes,omitempty"`
}

type SourceCatalogueRow struct {
	TraceID         string            `json:"trace_id"`
	RawName         string            `json:"raw_name,omitempty"`
	DisplayName     string            `json:"display_name,omitempty"`
	Unit            string            `json:"unit,omitempty"`
	ValueType       string            `json:"value_type,omitempty"`
	Access          string            `json:"access,omitempty"`
	GraphSource     string            `json:"graph_source,omitempty"`
	GraphType       string            `json:"graph_type,omitempty"`
	Category        SignalCategory    `json:"category,omitempty"`
	Kind            SignalKind        `json:"kind,omitempty"`
	Role            SignalRole        `json:"role,omitempty"`
	DefaultHint     GraphHint         `json:"default_hint,omitempty"`
	StaticInfo      bool              `json:"static_info,omitempty"`
	SemanticStatus  string            `json:"semantic_status,omitempty"`
	SourceSubject   string            `json:"source_subject,omitempty"`
	HistoryPath     string            `json:"history_path,omitempty"`
	TargetID        string            `json:"target_id,omitempty"`
	TargetFormat    string            `json:"target_format,omitempty"`
	TargetUse       string            `json:"target_use,omitempty"`
	OwnerKind       string            `json:"owner_kind,omitempty"`
	OwnerNodeID     string            `json:"owner_node_id,omitempty"`
	OwnerProcessID  int               `json:"owner_process_id,omitempty"`
	OwnershipMode   string            `json:"ownership_mode,omitempty"`
	DiscoveryBadges []string          `json:"discovery_badges,omitempty"`
	TargetMetadata  map[string]string `json:"target_metadata,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type DiscoveredCatalogueRecord struct {
	CatalogueID    string       `json:"catalogue_id"`
	LinkedSourceID string       `json:"linked_source_id,omitempty"`
	DisplayName    string       `json:"display_name"`
	SourceFamily   SourceFamily `json:"source_family"`
	Status         string       `json:"status"`
	Owner          string       `json:"owner,omitempty"`
	EntryCount     int          `json:"entry_count,omitempty"`
	RouteHint      string       `json:"route_hint,omitempty"`
	ExplorerGroup  string       `json:"explorer_group,omitempty"`
	Priority       string       `json:"priority,omitempty"`
	Notes          []string     `json:"notes,omitempty"`
}

type SourceSignalSelection struct {
	SchemaVersion int              `json:"schema_version"`
	SourceID      string           `json:"source_id"`
	SourceFamily  SourceFamily     `json:"source_family,omitempty"`
	Signals       []SelectedSignal `json:"signals"`
}

type SelectedSignal struct {
	SignalID       SignalID       `json:"signal_id"`
	TraceID        string         `json:"trace_id"`
	CanonicalName  string         `json:"canonical_name,omitempty"`
	DisplayName    string         `json:"display_name,omitempty"`
	Subsystem      string         `json:"subsystem,omitempty"`
	Category       SignalCategory `json:"category,omitempty"`
	Kind           SignalKind     `json:"kind,omitempty"`
	Unit           string         `json:"unit,omitempty"`
	Role           SignalRole     `json:"role,omitempty"`
	SourceInstance string         `json:"source_instance,omitempty"`
	DUTID          DUTID          `json:"dut_id,omitempty"`
	DefaultHint    GraphHint      `json:"default_hint,omitempty"`
	Availability   Availability   `json:"availability,omitempty"`
	StaticInfo     bool           `json:"static_info,omitempty"`
}

type ResolvedSignalSelection struct {
	Signals         []SemanticSignal `json:"signals"`
	UnselectedCount int              `json:"unselected_count"`
}

type GlobalSourceCatalogue struct {
	SchemaVersion        int                         `json:"schema_version"`
	GeneratedAt          string                      `json:"generated_at,omitempty"`
	SelectionOwner       string                      `json:"selection_owner,omitempty"`
	Catalogues           []SourceCatalogue           `json:"catalogues"`
	Selections           []SourceSignalSelection     `json:"selections,omitempty"`
	DiscoveryPolicy      SourceCapabilities          `json:"discovery_policy,omitempty"`
	DiscoveredCatalogues []DiscoveredCatalogueRecord `json:"discovered_catalogues,omitempty"`
}

type GlobalSourceCatalogueSummary struct {
	SourceCount              int                           `json:"source_count"`
	EntryCount               int                           `json:"entry_count"`
	SelectedCount            int                           `json:"selected_count"`
	UnselectedCount          int                           `json:"unselected_count"`
	DiscoveredCatalogueCount int                           `json:"discovered_catalogue_count,omitempty"`
	KnownAbsentCount         int                           `json:"known_absent_count,omitempty"`
	ByFamily                 map[SourceFamily]int          `json:"by_family"`
	Capabilities             map[string]SourceCapabilities `json:"capabilities,omitempty"`
}

func (c SourceCatalogue) Validate() error {
	if c.SchemaVersion != CurrentSourceCatalogueSchemaVersion {
		return fmt.Errorf("source catalogue schema_version must be %d, got %d", CurrentSourceCatalogueSchemaVersion, c.SchemaVersion)
	}
	if strings.TrimSpace(c.SourceID) == "" {
		return fmt.Errorf("source catalogue source_id is required")
	}
	if c.SourceFamily == "" {
		return fmt.Errorf("source catalogue source_family is required")
	}
	for i, path := range c.Capabilities.TransportPaths {
		if err := path.Validate(); err != nil {
			return fmt.Errorf("source catalogue transport path %d invalid: %w", i, err)
		}
	}
	seen := map[string]struct{}{}
	for i, entry := range c.Entries {
		traceID := strings.TrimSpace(entry.TraceID)
		if traceID == "" {
			return fmt.Errorf("source catalogue entry %d trace_id is required", i)
		}
		if _, ok := seen[traceID]; ok {
			return fmt.Errorf("source catalogue duplicate trace_id %q", traceID)
		}
		seen[traceID] = struct{}{}
		if strings.TrimSpace(entry.Access) == "" && !entry.StaticInfo {
			return fmt.Errorf("source catalogue entry %d access is required unless static_info is true", i)
		}
	}
	return nil
}

func (p TransportPath) Validate() error {
	if strings.TrimSpace(p.PathID) == "" {
		return fmt.Errorf("path_id is required")
	}
	if strings.TrimSpace(p.PathKind) == "" {
		return fmt.Errorf("path_kind is required")
	}
	if strings.TrimSpace(p.PhysicalTransport) == "" && strings.TrimSpace(p.NetworkTransport) == "" {
		return fmt.Errorf("physical_transport or network_transport is required")
	}
	if strings.TrimSpace(p.State) == "" {
		return fmt.Errorf("state is required")
	}
	return nil
}

func (d DiscoveredCatalogueRecord) Validate() error {
	if strings.TrimSpace(d.CatalogueID) == "" {
		return fmt.Errorf("catalogue_id is required")
	}
	if strings.TrimSpace(d.DisplayName) == "" {
		return fmt.Errorf("display_name is required")
	}
	if d.SourceFamily == "" {
		return fmt.Errorf("source_family is required")
	}
	if strings.TrimSpace(d.Status) == "" {
		return fmt.Errorf("status is required")
	}
	if strings.TrimSpace(d.RouteHint) != "" && !strings.HasPrefix(strings.TrimSpace(d.RouteHint), "/") {
		return fmt.Errorf("route_hint must be a local UI gateway path")
	}
	return nil
}

func (s SourceSignalSelection) ValidateAgainst(c SourceCatalogue) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if s.SchemaVersion != CurrentSourceCatalogueSchemaVersion {
		return fmt.Errorf("source signal selection schema_version must be %d, got %d", CurrentSourceCatalogueSchemaVersion, s.SchemaVersion)
	}
	if strings.TrimSpace(s.SourceID) == "" {
		return fmt.Errorf("source signal selection source_id is required")
	}
	if s.SourceID != c.SourceID {
		return fmt.Errorf("source signal selection source_id %q does not match catalogue source_id %q", s.SourceID, c.SourceID)
	}
	if s.SourceFamily != "" && s.SourceFamily != c.SourceFamily {
		return fmt.Errorf("source signal selection source_family %q does not match catalogue source_family %q", s.SourceFamily, c.SourceFamily)
	}
	entries := c.entriesByTraceID()
	seenSignalIDs := map[SignalID]struct{}{}
	seenTraceIDs := map[string]struct{}{}
	for i, signal := range s.Signals {
		if signal.SignalID == "" {
			return fmt.Errorf("source signal selection entry %d signal_id is required", i)
		}
		if strings.TrimSpace(signal.TraceID) == "" {
			return fmt.Errorf("source signal selection entry %d trace_id is required", i)
		}
		if _, ok := entries[signal.TraceID]; !ok {
			return fmt.Errorf("source signal selection entry %d trace_id %q is not in source catalogue", i, signal.TraceID)
		}
		if _, ok := seenSignalIDs[signal.SignalID]; ok {
			return fmt.Errorf("source signal selection duplicate signal_id %q", signal.SignalID)
		}
		if _, ok := seenTraceIDs[signal.TraceID]; ok {
			return fmt.Errorf("source signal selection duplicate trace_id %q", signal.TraceID)
		}
		seenSignalIDs[signal.SignalID] = struct{}{}
		seenTraceIDs[signal.TraceID] = struct{}{}
	}
	return nil
}

func ResolveSourceSignalSelection(c SourceCatalogue, s SourceSignalSelection) (ResolvedSignalSelection, error) {
	if err := s.ValidateAgainst(c); err != nil {
		return ResolvedSignalSelection{}, err
	}
	entries := c.entriesByTraceID()
	selectedTraceIDs := map[string]struct{}{}
	resolved := ResolvedSignalSelection{
		Signals: make([]SemanticSignal, 0, len(s.Signals)),
	}
	for _, selected := range s.Signals {
		entry := entries[selected.TraceID]
		selectedTraceIDs[selected.TraceID] = struct{}{}
		resolved.Signals = append(resolved.Signals, selected.resolve(c, entry))
	}
	for _, entry := range c.Entries {
		if _, ok := selectedTraceIDs[entry.TraceID]; !ok {
			resolved.UnselectedCount++
		}
	}
	return resolved, nil
}

func (g GlobalSourceCatalogue) Validate() error {
	if g.SchemaVersion != CurrentSourceCatalogueSchemaVersion {
		return fmt.Errorf("global source catalogue schema_version must be %d, got %d", CurrentSourceCatalogueSchemaVersion, g.SchemaVersion)
	}
	discovered := map[string]struct{}{}
	for _, record := range g.DiscoveredCatalogues {
		if err := record.Validate(); err != nil {
			return fmt.Errorf("discovered catalogue %q invalid: %w", record.CatalogueID, err)
		}
		if _, exists := discovered[record.CatalogueID]; exists {
			return fmt.Errorf("global source catalogue duplicate discovered catalogue_id %q", record.CatalogueID)
		}
		discovered[record.CatalogueID] = struct{}{}
	}
	catalogues := make(map[string]SourceCatalogue, len(g.Catalogues))
	for _, catalogue := range g.Catalogues {
		if err := catalogue.Validate(); err != nil {
			return err
		}
		if _, exists := catalogues[catalogue.SourceID]; exists {
			return fmt.Errorf("global source catalogue duplicate source_id %q", catalogue.SourceID)
		}
		catalogues[catalogue.SourceID] = catalogue
	}
	for _, record := range g.DiscoveredCatalogues {
		linkedSourceID := strings.TrimSpace(record.LinkedSourceID)
		if linkedSourceID == "" {
			continue
		}
		catalogue, ok := catalogues[linkedSourceID]
		if !ok {
			return fmt.Errorf("discovered catalogue %q linked_source_id %q has no catalogue", record.CatalogueID, linkedSourceID)
		}
		if catalogue.SourceFamily != record.SourceFamily {
			return fmt.Errorf("discovered catalogue %q linked_source_id %q family %q does not match %q", record.CatalogueID, linkedSourceID, catalogue.SourceFamily, record.SourceFamily)
		}
	}
	for _, selection := range g.Selections {
		catalogue, ok := catalogues[selection.SourceID]
		if !ok {
			return fmt.Errorf("source signal selection source_id %q has no catalogue", selection.SourceID)
		}
		if err := selection.ValidateAgainst(catalogue); err != nil {
			return err
		}
	}
	return nil
}

func (g GlobalSourceCatalogue) Summary() (GlobalSourceCatalogueSummary, error) {
	if err := g.Validate(); err != nil {
		return GlobalSourceCatalogueSummary{}, err
	}
	summary := GlobalSourceCatalogueSummary{
		SourceCount:              len(g.Catalogues),
		DiscoveredCatalogueCount: len(g.DiscoveredCatalogues),
		ByFamily:                 map[SourceFamily]int{},
		Capabilities:             map[string]SourceCapabilities{},
	}
	for _, record := range g.DiscoveredCatalogues {
		if record.Status == "known_absent" {
			summary.KnownAbsentCount++
		}
	}
	selectedBySource := make(map[string]int, len(g.Selections))
	for _, catalogue := range g.Catalogues {
		summary.EntryCount += len(catalogue.Entries)
		summary.ByFamily[catalogue.SourceFamily] += len(catalogue.Entries)
		summary.Capabilities[catalogue.SourceID] = catalogue.Capabilities
	}
	for _, selection := range g.Selections {
		summary.SelectedCount += len(selection.Signals)
		selectedBySource[selection.SourceID] += len(selection.Signals)
		catalogue := sourceCatalogueByID(g.Catalogues, selection.SourceID)
		if _, err := ResolveSourceSignalSelection(catalogue, selection); err != nil {
			return GlobalSourceCatalogueSummary{}, err
		}
	}
	for _, catalogue := range g.Catalogues {
		selected := selectedBySource[catalogue.SourceID]
		if selected < len(catalogue.Entries) {
			summary.UnselectedCount += len(catalogue.Entries) - selected
		}
	}
	return summary, nil
}

func (c SourceCatalogue) entriesByTraceID() map[string]SourceCatalogueRow {
	out := make(map[string]SourceCatalogueRow, len(c.Entries))
	for _, entry := range c.Entries {
		out[entry.TraceID] = entry
	}
	return out
}

func (s SelectedSignal) resolve(c SourceCatalogue, entry SourceCatalogueRow) SemanticSignal {
	availability := s.Availability
	if availability == "" {
		availability = AvailabilityDiscovered
	}
	if s.StaticInfo {
		availability = AvailabilityFixed
	}
	canonicalName := firstNonEmpty(s.CanonicalName, entry.RawName, entry.TraceID, string(s.SignalID))
	displayName := firstNonEmpty(s.DisplayName, entry.DisplayName, entry.RawName, entry.TraceID, string(s.SignalID))
	return SemanticSignal{
		SignalID:       s.SignalID,
		CanonicalName:  canonicalName,
		RawName:        firstNonEmpty(entry.RawName, entry.TraceID),
		DisplayName:    displayName,
		Subsystem:      s.Subsystem,
		Category:       firstCategory(s.Category, entry.Category),
		Kind:           firstKind(s.Kind, entry.Kind),
		Unit:           firstNonEmpty(s.Unit, entry.Unit),
		Role:           firstRole(s.Role, entry.Role),
		SourceFamily:   c.SourceFamily,
		Availability:   availability,
		SourceInstance: firstNonEmpty(s.SourceInstance, c.SourceID),
		DUTID:          s.DUTID,
		DefaultHint:    firstHint(s.DefaultHint, entry.DefaultHint),
	}
}

func sourceCatalogueByID(catalogues []SourceCatalogue, sourceID string) SourceCatalogue {
	for _, catalogue := range catalogues {
		if catalogue.SourceID == sourceID {
			return catalogue
		}
	}
	return SourceCatalogue{}
}

func firstCategory(values ...SignalCategory) SignalCategory {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstKind(values ...SignalKind) SignalKind {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstRole(values ...SignalRole) SignalRole {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstHint(values ...GraphHint) GraphHint {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
