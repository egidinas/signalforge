package graphsem

import (
	"fmt"
	"strings"
)

const CurrentSignalDictionarySchemaVersion = 2

type SignalDictionaryClassification string

const (
	SignalDictionaryPublicSynthetic SignalDictionaryClassification = "public_synthetic"
	SignalDictionaryPublicCleanRoom SignalDictionaryClassification = "public_clean_room"
	SignalDictionaryPrivateLab      SignalDictionaryClassification = "private_lab"
	SignalDictionaryDerived         SignalDictionaryClassification = "derived"
)

type SignalRouteOperation string

const (
	SignalRouteObserve SignalRouteOperation = "observe"
	SignalRouteMux     SignalRouteOperation = "mux"
	SignalRouteRoute   SignalRouteOperation = "route"
	SignalRouteMirror  SignalRouteOperation = "mirror"
	SignalRouteTap     SignalRouteOperation = "tap"
	SignalRouteSplice  SignalRouteOperation = "splice"
	SignalRouteReplay  SignalRouteOperation = "replay"
)

type SignalDictionaryBundle struct {
	SchemaVersion      int                            `json:"schema_version"`
	FixtureID          string                         `json:"fixture_id"`
	Classification     SignalDictionaryClassification `json:"classification"`
	GeneratedAt        string                         `json:"generated_at,omitempty"`
	DefinitionProfiles []SignalDefinitionProfile      `json:"definition_profiles,omitempty"`
	Catalogues         []SourceCatalogue              `json:"catalogues,omitempty"`
	Projections        []SignalProjectionBundle       `json:"projections,omitempty"`
	SemanticGroups     []SignalSemanticGroup          `json:"semantic_groups,omitempty"`
	Routes             []SignalRouteContract          `json:"routes,omitempty"`
	Rings              []SignalRingProfile            `json:"rings,omitempty"`
	Decimations        []SignalDecimationProfile      `json:"decimations,omitempty"`
	GraphWall          []SignalGraphWallTarget        `json:"graph_wall_targets,omitempty"`
	DomainMetadata     []SignalDomainMetadata         `json:"domain_metadata,omitempty"`
	Metadata           map[string]any                 `json:"metadata,omitempty"`
}

type SignalDefinitionProfile struct {
	ID             string                         `json:"id"`
	DisplayName    string                         `json:"display_name,omitempty"`
	System         string                         `json:"system"`
	Family         string                         `json:"family"`
	SubFamily      string                         `json:"sub_family,omitempty"`
	Variant        string                         `json:"variant,omitempty"`
	Version        string                         `json:"version,omitempty"`
	SourceFamilies []SourceFamily                 `json:"source_families,omitempty"`
	Description    string                         `json:"description,omitempty"`
	Classification SignalDictionaryClassification `json:"classification,omitempty"`
	Metadata       map[string]any                 `json:"metadata,omitempty"`
}

type SignalValueEncoding struct {
	Kind      string           `json:"kind"`
	DataType  string           `json:"data_type,omitempty"`
	ByteOrder string           `json:"byte_order,omitempty"`
	StartBit  int              `json:"start_bit,omitempty"`
	BitLength int              `json:"bit_length,omitempty"`
	Scale     float64          `json:"scale,omitempty"`
	Offset    float64          `json:"offset,omitempty"`
	RawUnit   string           `json:"raw_unit,omitempty"`
	BitFields []SignalBitField `json:"bit_fields,omitempty"`
	Metadata  map[string]any   `json:"metadata,omitempty"`
}

type SignalBitField struct {
	Name       string            `json:"name"`
	Label      string            `json:"label,omitempty"`
	Bit        int               `json:"bit,omitempty"`
	Mask       string            `json:"mask,omitempty"`
	ValueTable map[string]string `json:"value_table,omitempty"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
}

type SignalSemanticGroup struct {
	ID             string                `json:"id"`
	Label          string                `json:"label"`
	ParentID       string                `json:"parent_id,omitempty"`
	SourceFamilies []SourceFamily        `json:"source_families,omitempty"`
	SourceIDs      []string              `json:"source_ids,omitempty"`
	SignalRefs     []SignalProjectionRef `json:"signal_refs,omitempty"`
	SortKey        string                `json:"sort_key,omitempty"`
	DefaultOpen    bool                  `json:"default_open,omitempty"`
	Metadata       map[string]any        `json:"metadata,omitempty"`
}

type SignalRouteContract struct {
	RouteID           string                `json:"route_id"`
	Operation         SignalRouteOperation  `json:"operation"`
	Label             string                `json:"label,omitempty"`
	SourceEndpointID  string                `json:"source_endpoint_id,omitempty"`
	SinkEndpointID    string                `json:"sink_endpoint_id,omitempty"`
	TransportKind     string                `json:"transport_kind,omitempty"`
	AccessMode        string                `json:"access_mode"`
	AuthorityMode     string                `json:"authority_mode,omitempty"`
	LeaseRequired     bool                  `json:"lease_required,omitempty"`
	OperatorAck       bool                  `json:"operator_ack_required,omitempty"`
	RollbackAvailable bool                  `json:"rollback_available,omitempty"`
	FreshnessMS       int                   `json:"freshness_ms,omitempty"`
	InputRefs         []SignalProjectionRef `json:"input_refs,omitempty"`
	OutputRefs        []SignalProjectionRef `json:"output_refs,omitempty"`
	DBC               *SignalDBCMetadata    `json:"dbc,omitempty"`
	RingID            string                `json:"ring_id,omitempty"`
	DecimationID      string                `json:"decimation_id,omitempty"`
	Metadata          map[string]any        `json:"metadata,omitempty"`
}

type SignalDBCMetadata struct {
	BusID       string            `json:"bus_id"`
	MessageName string            `json:"message_name"`
	FrameID     string            `json:"frame_id"`
	ExtendedID  bool              `json:"extended_id,omitempty"`
	CANFD       bool              `json:"can_fd,omitempty"`
	DLC         int               `json:"dlc,omitempty"`
	CycleTimeMS int               `json:"cycle_time_ms,omitempty"`
	SendType    string            `json:"send_type,omitempty"`
	MuxSwitch   string            `json:"mux_switch,omitempty"`
	MuxCase     string            `json:"mux_case,omitempty"`
	MuxValue    string            `json:"mux_value,omitempty"`
	SignalName  string            `json:"signal_name,omitempty"`
	ValueTable  map[string]string `json:"value_table,omitempty"`
	Layout      *SignalDBCLayout  `json:"layout,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

type SignalDBCLayout struct {
	StartBit  int     `json:"start_bit"`
	BitLength int     `json:"bit_length"`
	ByteOrder string  `json:"byte_order"`
	Signed    bool    `json:"signed,omitempty"`
	Factor    float64 `json:"factor,omitempty"`
	Offset    float64 `json:"offset,omitempty"`
	Minimum   float64 `json:"minimum,omitempty"`
	Maximum   float64 `json:"maximum,omitempty"`
	RawUnit   string  `json:"raw_unit,omitempty"`
}

type SignalRingProfile struct {
	ID              string                `json:"id"`
	Label           string                `json:"label,omitempty"`
	SourceID        string                `json:"source_id,omitempty"`
	SignalRefs      []SignalProjectionRef `json:"signal_refs,omitempty"`
	CapacitySamples int                   `json:"capacity_samples,omitempty"`
	CapacityBytes   int                   `json:"capacity_bytes,omitempty"`
	RetentionPolicy string                `json:"retention_policy"`
	SequenceField   string                `json:"sequence_field,omitempty"`
	WatermarkField  string                `json:"watermark_field,omitempty"`
	DroppedField    string                `json:"dropped_field,omitempty"`
	TruncatedField  string                `json:"truncated_field,omitempty"`
	FreshnessMS     int                   `json:"freshness_ms,omitempty"`
	Metadata        map[string]any        `json:"metadata,omitempty"`
}

type SignalDecimationProfile struct {
	ID                  string         `json:"id"`
	Label               string         `json:"label,omitempty"`
	Algorithm           string         `json:"algorithm"`
	AppliesToKinds      []SignalKind   `json:"applies_to_kinds,omitempty"`
	MaxPoints           int            `json:"max_points,omitempty"`
	BucketMS            int            `json:"bucket_ms,omitempty"`
	MinMaxEnvelope      bool           `json:"min_max_envelope,omitempty"`
	EventPreserving     bool           `json:"event_preserving,omitempty"`
	StateSpanPreserving bool           `json:"state_span_preserving,omitempty"`
	ReducerProvenance   string         `json:"reducer_provenance,omitempty"`
	Metadata            map[string]any `json:"metadata,omitempty"`
}

type SignalGraphWallTarget struct {
	TargetID     string                `json:"target_id"`
	Label        string                `json:"label,omitempty"`
	Lane         string                `json:"lane"`
	Role         string                `json:"role"`
	TileLevel    string                `json:"tile_level,omitempty"`
	SourceID     string                `json:"source_id,omitempty"`
	SourceFamily SourceFamily          `json:"source_family,omitempty"`
	TraceID      string                `json:"trace_id,omitempty"`
	ProjectionID string                `json:"projection_id,omitempty"`
	RouteID      string                `json:"route_id,omitempty"`
	RingID       string                `json:"ring_id,omitempty"`
	DecimationID string                `json:"decimation_id,omitempty"`
	SignalRefs   []SignalProjectionRef `json:"signal_refs,omitempty"`
	Metadata     map[string]any        `json:"metadata,omitempty"`
}

type SignalDomainMetadata struct {
	Domain        SourceFamily   `json:"domain,omitempty"`
	SourceID      string         `json:"source_id,omitempty"`
	DefinitionRef string         `json:"definition_ref,omitempty"`
	System        string         `json:"system,omitempty"`
	Family        string         `json:"family,omitempty"`
	SubFamily     string         `json:"sub_family,omitempty"`
	Required      []string       `json:"required,omitempty"`
	Recommended   []string       `json:"recommended,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

func ValidateSignalDictionaryBundle(bundle SignalDictionaryBundle) error {
	if bundle.SchemaVersion != CurrentSignalDictionarySchemaVersion {
		return fmt.Errorf("signal dictionary schema_version must be %d, got %d", CurrentSignalDictionarySchemaVersion, bundle.SchemaVersion)
	}
	if strings.TrimSpace(bundle.FixtureID) == "" {
		return fmt.Errorf("signal dictionary fixture_id is required")
	}
	if bundle.Classification == "" {
		return fmt.Errorf("signal dictionary classification is required")
	}
	definitions, err := validateSignalDefinitionProfiles(bundle.DefinitionProfiles)
	if err != nil {
		return err
	}
	index, err := newSignalDictionaryIndex(bundle.Catalogues)
	if err != nil {
		return err
	}
	if err := validateCatalogueDefinitionRefs(bundle.Catalogues, definitions); err != nil {
		return err
	}
	for i, projection := range bundle.Projections {
		if err := ValidateSignalProjectionBundle(projection, SignalProjectionValidationOptions{Catalogues: bundle.Catalogues}); err != nil {
			return fmt.Errorf("projection %d invalid: %w", i, err)
		}
	}
	seenGroups := map[string]struct{}{}
	for i, group := range bundle.SemanticGroups {
		if strings.TrimSpace(group.ID) == "" {
			return fmt.Errorf("semantic group %d id is required", i)
		}
		if strings.TrimSpace(group.Label) == "" {
			return fmt.Errorf("semantic group %q label is required", group.ID)
		}
		if _, exists := seenGroups[group.ID]; exists {
			return fmt.Errorf("duplicate semantic group %q", group.ID)
		}
		seenGroups[group.ID] = struct{}{}
		if group.ParentID != "" {
			if _, ok := seenGroups[group.ParentID]; !ok {
				return fmt.Errorf("semantic group %q parent_id %q must reference an earlier group", group.ID, group.ParentID)
			}
		}
		if err := index.validateRefs(group.SignalRefs); err != nil {
			return fmt.Errorf("semantic group %q invalid: %w", group.ID, err)
		}
	}
	routeIDs := map[string]struct{}{}
	for i, route := range bundle.Routes {
		if err := index.validateRoute(route); err != nil {
			return fmt.Errorf("route %d invalid: %w", i, err)
		}
		if _, exists := routeIDs[route.RouteID]; exists {
			return fmt.Errorf("duplicate route_id %q", route.RouteID)
		}
		routeIDs[route.RouteID] = struct{}{}
	}
	ringIDs := map[string]struct{}{}
	for i, ring := range bundle.Rings {
		if err := index.validateRing(ring); err != nil {
			return fmt.Errorf("ring %d invalid: %w", i, err)
		}
		if _, exists := ringIDs[ring.ID]; exists {
			return fmt.Errorf("duplicate ring id %q", ring.ID)
		}
		ringIDs[ring.ID] = struct{}{}
	}
	decimationIDs := map[string]struct{}{}
	for i, decimation := range bundle.Decimations {
		if err := validateDecimation(decimation); err != nil {
			return fmt.Errorf("decimation %d invalid: %w", i, err)
		}
		if _, exists := decimationIDs[decimation.ID]; exists {
			return fmt.Errorf("duplicate decimation id %q", decimation.ID)
		}
		decimationIDs[decimation.ID] = struct{}{}
	}
	for _, route := range bundle.Routes {
		if route.RingID != "" {
			if _, ok := ringIDs[route.RingID]; !ok {
				return fmt.Errorf("route %q references unknown ring_id %q", route.RouteID, route.RingID)
			}
		}
		if route.DecimationID != "" {
			if _, ok := decimationIDs[route.DecimationID]; !ok {
				return fmt.Errorf("route %q references unknown decimation_id %q", route.RouteID, route.DecimationID)
			}
		}
	}
	for i, target := range bundle.GraphWall {
		if err := index.validateGraphWallTarget(target, routeIDs, ringIDs, decimationIDs); err != nil {
			return fmt.Errorf("graph wall target %d invalid: %w", i, err)
		}
	}
	for i, domain := range bundle.DomainMetadata {
		if strings.TrimSpace(domain.DefinitionRef) != "" {
			if len(definitions) == 0 {
				return fmt.Errorf("domain metadata %d definition_ref %q has no definition_profiles table", i, domain.DefinitionRef)
			}
			if _, ok := definitions[domain.DefinitionRef]; !ok {
				return fmt.Errorf("domain metadata %d references unknown definition_ref %q", i, domain.DefinitionRef)
			}
		}
		if domain.Domain == "" && strings.TrimSpace(domain.DefinitionRef) == "" && strings.TrimSpace(domain.Family) == "" {
			return fmt.Errorf("domain metadata %d domain, definition_ref, or family is required", i)
		}
		if domain.SourceID != "" {
			catalogue, ok := index.catalogues[domain.SourceID]
			if !ok {
				return fmt.Errorf("domain metadata %d references unknown source_id %q", i, domain.SourceID)
			}
			if domain.Domain != "" && catalogue.SourceFamily != domain.Domain {
				return fmt.Errorf("domain metadata %d source %q family %q does not match domain %q", i, domain.SourceID, catalogue.SourceFamily, domain.Domain)
			}
		}
	}
	return nil
}

func validateSignalDefinitionProfiles(profiles []SignalDefinitionProfile) (map[string]SignalDefinitionProfile, error) {
	definitions := map[string]SignalDefinitionProfile{}
	for i, profile := range profiles {
		id := strings.TrimSpace(profile.ID)
		if id == "" {
			return definitions, fmt.Errorf("definition profile %d id is required", i)
		}
		if strings.TrimSpace(profile.System) == "" {
			return definitions, fmt.Errorf("definition profile %q system is required", id)
		}
		if strings.TrimSpace(profile.Family) == "" {
			return definitions, fmt.Errorf("definition profile %q family is required", id)
		}
		if _, exists := definitions[id]; exists {
			return definitions, fmt.Errorf("duplicate definition profile %q", id)
		}
		definitions[id] = profile
	}
	return definitions, nil
}

func validateCatalogueDefinitionRefs(catalogues []SourceCatalogue, definitions map[string]SignalDefinitionProfile) error {
	for i, catalogue := range catalogues {
		if strings.TrimSpace(catalogue.DefinitionRef) != "" {
			if len(definitions) == 0 {
				return fmt.Errorf("catalogue %d definition_ref %q has no definition_profiles table", i, catalogue.DefinitionRef)
			}
			if _, ok := definitions[catalogue.DefinitionRef]; !ok {
				return fmt.Errorf("catalogue %d references unknown definition_ref %q", i, catalogue.DefinitionRef)
			}
		}
		for j, row := range catalogue.Entries {
			if strings.TrimSpace(row.DefinitionRef) == "" {
				continue
			}
			if len(definitions) == 0 {
				return fmt.Errorf("catalogue %q entry %d definition_ref %q has no definition_profiles table", catalogue.SourceID, j, row.DefinitionRef)
			}
			if _, ok := definitions[row.DefinitionRef]; !ok {
				return fmt.Errorf("catalogue %q entry %d references unknown definition_ref %q", catalogue.SourceID, j, row.DefinitionRef)
			}
		}
	}
	return nil
}

type signalDictionaryIndex struct {
	catalogues map[string]SourceCatalogue
	traceRefs  map[string]struct{}
	targetIDs  map[string]struct{}
}

func newSignalDictionaryIndex(catalogues []SourceCatalogue) (signalDictionaryIndex, error) {
	index := signalDictionaryIndex{
		catalogues: map[string]SourceCatalogue{},
		traceRefs:  map[string]struct{}{},
		targetIDs:  map[string]struct{}{},
	}
	for _, catalogue := range catalogues {
		if err := catalogue.Validate(); err != nil {
			return index, err
		}
		if _, exists := index.catalogues[catalogue.SourceID]; exists {
			return index, fmt.Errorf("duplicate source_id %q", catalogue.SourceID)
		}
		index.catalogues[catalogue.SourceID] = catalogue
		for _, entry := range catalogue.Entries {
			addProjectionTraceKeys(index.traceRefs, catalogue.SourceFamily, catalogue.SourceID, entry.TraceID)
			if strings.TrimSpace(entry.TargetID) != "" {
				index.targetIDs[entry.TargetID] = struct{}{}
			}
		}
	}
	return index, nil
}

func (index signalDictionaryIndex) validateRefs(refs []SignalProjectionRef) error {
	for _, ref := range refs {
		key := SignalProjectionRefKey(ref)
		if key == "" {
			return fmt.Errorf("empty signal reference")
		}
		switch {
		case strings.HasPrefix(key, "trace:") && len(index.traceRefs) > 0:
			if _, ok := index.traceRefs[key]; !ok {
				return fmt.Errorf("unknown source trace %s", key)
			}
		case strings.HasPrefix(key, "target:") && len(index.targetIDs) > 0:
			if _, ok := index.targetIDs[strings.TrimPrefix(key, "target:")]; !ok {
				return fmt.Errorf("unknown target %s", key)
			}
		}
	}
	return nil
}

func (index signalDictionaryIndex) validateRoute(route SignalRouteContract) error {
	if strings.TrimSpace(route.RouteID) == "" {
		return fmt.Errorf("route_id is required")
	}
	if !validSignalRouteOperation(route.Operation) {
		return fmt.Errorf("route %q operation %q is not supported", route.RouteID, route.Operation)
	}
	if strings.TrimSpace(route.AccessMode) == "" {
		return fmt.Errorf("route %q access_mode is required", route.RouteID)
	}
	if err := index.validateRefs(route.InputRefs); err != nil {
		return fmt.Errorf("route %q input refs invalid: %w", route.RouteID, err)
	}
	if err := index.validateRefs(route.OutputRefs); err != nil {
		return fmt.Errorf("route %q output refs invalid: %w", route.RouteID, err)
	}
	if route.DBC != nil {
		if err := validateDBCMetadata(*route.DBC, route.Operation); err != nil {
			return fmt.Errorf("route %q dbc invalid: %w", route.RouteID, err)
		}
	}
	if isActiveSignalRouteOperation(route.Operation) && route.AccessMode != "read_only" && route.AccessMode != "dry_run" && !route.LeaseRequired {
		return fmt.Errorf("route %q active access requires lease_required", route.RouteID)
	}
	return nil
}

func (index signalDictionaryIndex) validateRing(ring SignalRingProfile) error {
	if strings.TrimSpace(ring.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if ring.CapacitySamples <= 0 && ring.CapacityBytes <= 0 {
		return fmt.Errorf("ring %q requires capacity_samples or capacity_bytes", ring.ID)
	}
	if strings.TrimSpace(ring.RetentionPolicy) == "" {
		return fmt.Errorf("ring %q retention_policy is required", ring.ID)
	}
	if strings.TrimSpace(ring.SourceID) != "" {
		if _, ok := index.catalogues[ring.SourceID]; !ok {
			return fmt.Errorf("ring %q references unknown source_id %q", ring.ID, ring.SourceID)
		}
	}
	if err := index.validateRefs(ring.SignalRefs); err != nil {
		return fmt.Errorf("ring %q signal refs invalid: %w", ring.ID, err)
	}
	return nil
}

func (index signalDictionaryIndex) validateGraphWallTarget(target SignalGraphWallTarget, routeIDs, ringIDs, decimationIDs map[string]struct{}) error {
	if strings.TrimSpace(target.TargetID) == "" {
		return fmt.Errorf("target_id is required")
	}
	if strings.TrimSpace(target.Lane) == "" {
		return fmt.Errorf("target %q lane is required", target.TargetID)
	}
	if strings.TrimSpace(target.Role) == "" {
		return fmt.Errorf("target %q role is required", target.TargetID)
	}
	if target.SourceID != "" {
		catalogue, ok := index.catalogues[target.SourceID]
		if !ok {
			return fmt.Errorf("target %q references unknown source_id %q", target.TargetID, target.SourceID)
		}
		if target.SourceFamily != "" && target.SourceFamily != catalogue.SourceFamily {
			return fmt.Errorf("target %q source_family %q does not match source %q family %q", target.TargetID, target.SourceFamily, target.SourceID, catalogue.SourceFamily)
		}
		if target.TraceID != "" {
			ref := SignalProjectionRef{SourceID: target.SourceID, SourceFamily: catalogue.SourceFamily, TraceID: target.TraceID}
			if err := index.validateRefs([]SignalProjectionRef{ref}); err != nil {
				return fmt.Errorf("target %q trace invalid: %w", target.TargetID, err)
			}
		}
	}
	if err := index.validateRefs(target.SignalRefs); err != nil {
		return fmt.Errorf("target %q signal refs invalid: %w", target.TargetID, err)
	}
	if target.RouteID != "" {
		if _, ok := routeIDs[target.RouteID]; !ok {
			return fmt.Errorf("target %q references unknown route_id %q", target.TargetID, target.RouteID)
		}
	}
	if target.RingID != "" {
		if _, ok := ringIDs[target.RingID]; !ok {
			return fmt.Errorf("target %q references unknown ring_id %q", target.TargetID, target.RingID)
		}
	}
	if target.DecimationID != "" {
		if _, ok := decimationIDs[target.DecimationID]; !ok {
			return fmt.Errorf("target %q references unknown decimation_id %q", target.TargetID, target.DecimationID)
		}
	}
	return nil
}

func validSignalRouteOperation(operation SignalRouteOperation) bool {
	switch operation {
	case SignalRouteObserve, SignalRouteMux, SignalRouteRoute, SignalRouteMirror, SignalRouteTap, SignalRouteSplice, SignalRouteReplay:
		return true
	default:
		return false
	}
}

func isActiveSignalRouteOperation(operation SignalRouteOperation) bool {
	switch operation {
	case SignalRouteMux, SignalRouteRoute, SignalRouteMirror, SignalRouteSplice:
		return true
	default:
		return false
	}
}

func validateDBCMetadata(dbc SignalDBCMetadata, operation SignalRouteOperation) error {
	if strings.TrimSpace(dbc.BusID) == "" {
		return fmt.Errorf("bus_id is required")
	}
	if strings.TrimSpace(dbc.MessageName) == "" {
		return fmt.Errorf("message_name is required")
	}
	if strings.TrimSpace(dbc.FrameID) == "" {
		return fmt.Errorf("frame_id is required")
	}
	if operation == SignalRouteMux {
		if strings.TrimSpace(dbc.MuxSwitch) == "" {
			return fmt.Errorf("mux_switch is required for mux routes")
		}
		if strings.TrimSpace(dbc.MuxCase) == "" && strings.TrimSpace(dbc.MuxValue) == "" {
			return fmt.Errorf("mux_case or mux_value is required for mux routes")
		}
	}
	if dbc.Layout != nil {
		if dbc.Layout.BitLength <= 0 {
			return fmt.Errorf("layout bit_length must be positive")
		}
		if strings.TrimSpace(dbc.Layout.ByteOrder) == "" {
			return fmt.Errorf("layout byte_order is required")
		}
	}
	return nil
}

func validateDecimation(decimation SignalDecimationProfile) error {
	if strings.TrimSpace(decimation.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if strings.TrimSpace(decimation.Algorithm) == "" {
		return fmt.Errorf("decimation %q algorithm is required", decimation.ID)
	}
	if decimation.MaxPoints <= 0 && decimation.BucketMS <= 0 {
		return fmt.Errorf("decimation %q requires max_points or bucket_ms", decimation.ID)
	}
	return nil
}
