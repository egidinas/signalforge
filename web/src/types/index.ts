// SignalForge shared type contracts — graph_tile.v1 + render layer types

import type {
  SignalProjectionBundle,
  SignalProjectionSignalRef,
  SignalProjectionSourceFamily,
} from "../catalogue/projection";

// ---- Render layer types ----

export type GraphPoint = { timestamp: string; value: number };

export type GraphTimeAxis = {
  start: string; end: string; anchor: string; now?: string;
  range_seconds: number; clamp: boolean; latest_policy: string;
  default_window_start?: string; default_window_end?: string;
};

export type GraphYAxis = {
  id: string; label: string; units: string; scale: string;
  min: number; max: number; side: string; format: string;
};

export type GraphTrace = {
  id: string; label: string; role: string; units: string;
  axis_id: string; source: string; values: GraphPoint[];
};

export type CompanionGraphGroup = {
  id: string; label: string; axes: GraphYAxis[]; traces: GraphTrace[];
};

export type HeroGraphModel = {
  id: string; title: string; owner: string; provenance: string;
  time_axis: GraphTimeAxis;
  axes: GraphYAxis[]; traces: GraphTrace[];
  phase_bands: GraphBand[]; dwell_windows: GraphBand[];
  markers: GraphMarker[]; companion_groups: CompanionGraphGroup[];
  execution?: { now?: string };
};

export type GraphWallSignal = {
  id: string; label: string; unit?: string; source: string;
  source_family: string; kind: string; category: string;
  role: string; subsystem: string; color?: string; axis_id?: string; section_id: string;
  value_table?: Record<string, string>;
};

export type GraphCardPlacement = {
  section_id: string; group_id: string; order: number;
  height_weight: number; default_visible: boolean; pinned: boolean;
  colocated_with?: string; resize_policy: string;
};

export type GraphWallCard = {
  id: string; title: string;
  kind: "line" | "counter" | "state" | "event" | string;
  role: string; placement: GraphCardPlacement; transport: string;
  direction: string; unit?: string; axis_policy: string;
  source_family: string; overview?: boolean; bucket?: string;
  note?: string; render_kind?: string; tile_endpoint?: string;
  latest_endpoint?: string; collapsible?: boolean;
  default_expanded?: boolean; supports_time_zoom?: boolean;
  supports_y_zoom?: boolean; include_markers?: boolean;
  signals: GraphWallSignal[];
};

export type GraphWallTimeRange = {
  start: string; end: string; anchor: string; range_seconds: number;
  mode: string; source: string;
};

export type GraphTilePolicy = {
  default_points: number; max_points: number;
  live_tile_min_refresh_ms: number; history_tile_max_count: number;
  viewport_prefetch_px: number; tile_buffer_max_entries: number;
  tile_buffer_ttl_ms: number; resolution_levels: string[];
  subscriber_role: string; shared_timebase_required: boolean;
  legend_may_affect_plot_width: boolean; malformed_svg_path_hard_failure: boolean;
};

export type GraphSection = {
  id: string; title: string; group_id: string; transport: string;
  direction: string; status: string; unplotted_count: number;
  cards: GraphWallCard[];
};

export type GraphWallModel = {
  id: string; title: string; generated_at: string; source_mode: string;
  graph_version: string; owner: string; provenance: string;
  time_range: GraphWallTimeRange; tile_policy: GraphTilePolicy;
  graph_groups: unknown[]; sections: GraphSection[];
};

export type GraphTileCardRef = {
  card_id: string; section_id?: string; title: string; render_kind: string;
  unit?: string; axis_policy: string; tile_endpoint: string;
  latest_endpoint: string; tile_files?: unknown[]; default_expanded: boolean;
  collapsible: boolean; supports_time_zoom: boolean; supports_y_zoom: boolean;
  include_markers?: boolean; signals: GraphWallSignal[]; evidence_links?: unknown[];
};

// ---- Public semantic contracts ----

export type LifecycleState = "current" | "staged" | "prospective" | "submitted" | "confirmed" | "failed";

export type LifecycleStep = {
  state: LifecycleState;
  label?: string;
  value?: string | number | boolean | null;
  unit?: string;
  timestamp?: string;
  source?: string;
  evidence_ref?: string;
  note?: string;
  safe_to_apply?: boolean;
  counterpart_ids?: string[];
  meta?: Record<string, unknown>;
};

export type WriteLifecycleContract = {
  subject_id: string;
  subject_label?: string;
  target_id?: string;
  current?: LifecycleStep;
  staged?: LifecycleStep;
  prospective?: LifecycleStep;
  submission?: LifecycleStep;
  confirmed?: LifecycleStep;
  failed?: LifecycleStep & { error?: string; retryable?: boolean };
  steps?: LifecycleStep[];
  lease_token?: string;
  request_id?: string;
  transport?: string;
  endpoint?: string;
  safety_note?: string;
  evidence_ref?: string;
  counterpart_ids?: string[];
  meta?: Record<string, unknown>;
};

export type SemanticValueQuality = "unknown" | "good" | "stale" | "degraded" | "invalid" | "unavailable";

export type SemanticValueCounterpart = {
  id: string;
  label?: string;
  role?: string;
  value?: string | number | boolean | null;
  unit?: string;
  quality?: SemanticValueQuality;
  source?: string;
};

export type SemanticValueHoverMeta = {
  id?: string;
  label: string;
  value?: string | number | boolean | null;
  unit?: string;
  quality?: SemanticValueQuality;
  timestamp?: string;
  source?: string;
  evidence_ref?: string;
  help?: string;
  safety_note?: string;
  counterparts?: SemanticValueCounterpart[];
  role?: string;
  format?: string;
  precision?: number;
  meta?: Record<string, unknown>;
};

export type RouteState = "hot" | "warm" | "fallback" | "offline" | "unknown";

export type RouteCandidate = {
  id: string;
  label?: string;
  transport?: string;
  state: RouteState;
  endpoint?: string;
  latency_ms?: number;
  latency_unit?: "ms";
  rate?: number;
  rate_unit?: string;
  priority?: number;
  healthy?: boolean;
  preferred?: boolean;
  notes?: string;
  meta?: Record<string, unknown>;
};

export type RouteRedundancyContract = {
  subject_id: string;
  subject_label?: string;
  active_route_id?: string;
  hot?: RouteCandidate[];
  warm?: RouteCandidate[];
  fallback?: RouteCandidate[];
  candidates?: RouteCandidate[];
  transport?: string;
  endpoint?: string;
  redundancy_mode?: string;
  primary_state?: RouteState;
  meta?: Record<string, unknown>;
};

export type JsonSignalCatalogueEntry = SemanticSignal & {
  title?: string;
  description?: string;
  help?: string;
  safety_note?: string;
  evidence_ref?: string;
  counterparts?: SemanticValueCounterpart[];
};

export type JsonSignalCatalogue = {
  schema_version?: string | number;
  source?: string;
  generated_at?: string;
  signals: JsonSignalCatalogueEntry[];
  channels?: Channel[];
  meta?: Record<string, unknown>;
};

export type SourceCatalogueTransportPath = {
  path_id: string;
  path_kind: string;
  physical_transport?: string;
  network_transport?: string;
  endpoint?: string;
  state?: string;
  workflow?: string;
  notes?: string[];
};

export type SourceCatalogueCapabilities = {
  supports_live?: boolean;
  supports_history?: boolean;
  supports_metadata_only?: boolean;
  max_signals?: number;
  max_bytes_per_second?: number;
  default_rate_hz?: number;
  recommended_rate_hz?: number;
  discovery_cadence_sec?: number;
  subscription_endpoint?: string;
  history_endpoint?: string;
  live_subjects?: string[];
  selection_required?: boolean;
  polite_access_statement?: string;
  transport_paths?: SourceCatalogueTransportPath[];
};

export type SignalDefinitionProfile = {
  id: string;
  display_name?: string;
  system: string;
  family: string;
  sub_family?: string;
  variant?: string;
  version?: string;
  source_families?: SignalProjectionSourceFamily[];
  description?: string;
  classification?: SignalDictionaryClassification;
  metadata?: Record<string, unknown>;
};

export type SignalBitField = {
  name: string;
  label?: string;
  bit?: number;
  mask?: string;
  value_table?: Record<string, string>;
  metadata?: Record<string, unknown>;
};

export type SignalValueEncoding = {
  kind: string;
  data_type?: string;
  byte_order?: string;
  start_bit?: number;
  bit_length?: number;
  scale?: number;
  offset?: number;
  raw_unit?: string;
  bit_fields?: SignalBitField[];
  metadata?: Record<string, unknown>;
};

export type SourceCatalogueEntry = {
  trace_id: string;
  raw_name?: string;
  display_name?: string;
  description?: string;
  help?: string;
  safety_note?: string;
  source_evidence?: string[];
  unit?: string;
  value_type?: string;
  value_table?: Record<string, string>;
  encoding?: SignalValueEncoding;
  access?: string;
  graph_source?: string;
  graph_type?: string;
  category?: string;
  kind?: string;
  role?: string;
  group_key?: string;
  group_label?: string;
  instance_key?: string;
  sort_key?: string;
  counterpart_group?: string;
  counterpart_trace_ids?: string[];
  default_hint?: string;
  static_info?: boolean;
  semantic_status?: string;
  source_subject?: string;
  history_path?: string;
  target_id?: string;
  target_format?: string;
  target_use?: string;
  definition_ref?: string;
  owner_kind?: string;
  owner_node_id?: string;
  owner_process_id?: number;
  ownership_mode?: string;
  discovery_badges?: string[];
  target_metadata?: Record<string, string>;
  metadata?: Record<string, string>;
};

export type SourceCatalogue = {
  schema_version: number;
  source_id: string;
  source_family: string;
  display_name?: string;
  definition_ref?: string;
  entries: SourceCatalogueEntry[];
  capabilities?: SourceCatalogueCapabilities;
};

export type SignalDictionaryClassification = "public_synthetic" | "public_clean_room" | "private_lab" | "derived" | string;

export type SignalRouteOperation = "observe" | "mux" | "route" | "mirror" | "tap" | "splice" | "replay" | string;

export type SignalSemanticGroup = {
  id: string;
  label: string;
  parent_id?: string;
  source_families?: SignalProjectionSourceFamily[];
  source_ids?: string[];
  signal_refs?: SignalProjectionSignalRef[];
  sort_key?: string;
  default_open?: boolean;
  metadata?: Record<string, unknown>;
};

export type SignalDBCLayout = {
  start_bit: number;
  bit_length: number;
  byte_order: string;
  signed?: boolean;
  factor?: number;
  offset?: number;
  minimum?: number;
  maximum?: number;
  raw_unit?: string;
};

export type SignalDBCMetadata = {
  bus_id: string;
  message_name: string;
  frame_id: string;
  extended_id?: boolean;
  can_fd?: boolean;
  dlc?: number;
  cycle_time_ms?: number;
  send_type?: string;
  mux_switch?: string;
  mux_case?: string;
  mux_value?: string;
  signal_name?: string;
  value_table?: Record<string, string>;
  layout?: SignalDBCLayout;
  metadata?: Record<string, unknown>;
};

export type SignalRouteContract = {
  route_id: string;
  operation: SignalRouteOperation;
  label?: string;
  source_endpoint_id?: string;
  sink_endpoint_id?: string;
  transport_kind?: string;
  access_mode: string;
  authority_mode?: string;
  lease_required?: boolean;
  operator_ack_required?: boolean;
  rollback_available?: boolean;
  freshness_ms?: number;
  input_refs?: SignalProjectionSignalRef[];
  output_refs?: SignalProjectionSignalRef[];
  dbc?: SignalDBCMetadata;
  ring_id?: string;
  decimation_id?: string;
  metadata?: Record<string, unknown>;
};

export type SignalRingProfile = {
  id: string;
  label?: string;
  source_id?: string;
  signal_refs?: SignalProjectionSignalRef[];
  capacity_samples?: number;
  capacity_bytes?: number;
  retention_policy: string;
  sequence_field?: string;
  watermark_field?: string;
  dropped_field?: string;
  truncated_field?: string;
  freshness_ms?: number;
  metadata?: Record<string, unknown>;
};

export type SignalDecimationProfile = {
  id: string;
  label?: string;
  algorithm: string;
  applies_to_kinds?: string[];
  max_points?: number;
  bucket_ms?: number;
  min_max_envelope?: boolean;
  event_preserving?: boolean;
  state_span_preserving?: boolean;
  reducer_provenance?: string;
  metadata?: Record<string, unknown>;
};

export type SignalGraphWallTarget = {
  target_id: string;
  label?: string;
  lane: string;
  role: string;
  tile_level?: string;
  source_id?: string;
  source_family?: SignalProjectionSourceFamily;
  trace_id?: string;
  projection_id?: string;
  route_id?: string;
  ring_id?: string;
  decimation_id?: string;
  signal_refs?: SignalProjectionSignalRef[];
  metadata?: Record<string, unknown>;
};

export type SignalDomainMetadata = {
  domain?: SignalProjectionSourceFamily;
  source_id?: string;
  definition_ref?: string;
  system?: string;
  family?: string;
  sub_family?: string;
  required?: string[];
  recommended?: string[];
  metadata?: Record<string, unknown>;
};

export type SignalDictionaryBundle = {
  schema_version: number;
  fixture_id: string;
  classification: SignalDictionaryClassification;
  generated_at?: string;
  definition_profiles?: SignalDefinitionProfile[];
  catalogues?: SourceCatalogue[];
  projections?: SignalProjectionBundle[];
  semantic_groups?: SignalSemanticGroup[];
  routes?: SignalRouteContract[];
  rings?: SignalRingProfile[];
  decimations?: SignalDecimationProfile[];
  graph_wall_targets?: SignalGraphWallTarget[];
  domain_metadata?: SignalDomainMetadata[];
  metadata?: Record<string, unknown>;
};

// User semantic overlays keep operator labels, notes, and local projections
// separate from canonical signal catalogues.

export type SemanticOverlayTarget = {
  target_id?: string;
  device_id?: string;
  serial?: string;
  channel?: string | number;
  instance?: string | number;
  signal_id?: string | number;
  kind?: string;
  group?: string;
};

export type SemanticOverlayEntry = {
  id?: string;
  target: SemanticOverlayTarget;
  alias?: string;
  label?: string;
  note?: string;
  fixture_note?: string;
  hidden?: boolean;
  tags?: string[];
  source?: string;
  author?: string;
  updated_at?: string;
  meta?: Record<string, unknown>;
};

export type SemanticOverlayBundle = {
  schema_version?: string | number;
  namespace?: string;
  generated_at?: string;
  entries: SemanticOverlayEntry[];
  meta?: Record<string, unknown>;
};

export interface SemanticOverlayStoreOptions {
  namespace: string;
  storage?: Storage;
}

// ---- graph_tile.v1 types ----

export type TilePoint = {
  timestamp: string;
  value: number;
};

export type TileSpan = {
  start: string;
  end: string;
  value?: number;
  state?: string;
  label?: string;
  severity?: string;
};

export type TileSeries = {
  id: string;
  label: string;
  role: string;
  color?: string;
  unit?: string;
  units?: string;
  kind?: string;
  axis_id?: string;
  source: string;
  source_family?: string;
  render_kind?: string;
  step?: boolean;
  value_table?: Record<string, string>;
  points?: TilePoint[];
  spans?: TileSpan[];
};

export type TileDiagnostics = {
  source?: string;
  mode?: string;
  level?: string;
  requested_t0?: string;
  requested_t1?: string;
  raw_point_count?: number;
  point_count?: number;
  decimated?: boolean;
  decimation?: string;
  time_span_ms?: number;
  freshness_ms?: number;
  source_quality?: string;
  status?: string;
  renderer?: string;
  series_count?: number;
};

export type TileProvenance = {
  source?: string;
  source_node?: string;
  source_family?: string;
  mode?: string;
  generated_at?: string;
  synthetic?: boolean;
};

export type TileBand = {
  id: string;
  label?: string;
  kind?: string;
  start: string;
  end: string;
  series_a?: string;
  series_b?: string;
  fill?: string;
};

// Graph-model band (phase_band / dwell_window — start/end timestamps)
export type GraphBand = {
  id: string;
  label?: string;
  kind?: string;
  start: string;
  end: string;
  cycle_index?: number;
  target_deg_c?: number;
  result?: string;
};

export type GraphMarker = {
  id: string;
  label: string;
  timestamp: string;
  kind: string;
  role?: string;
  axis_id?: string;
  result?: string;
  requirement_id?: string;
  cycle_index?: number;
  value?: number;
  severity?: string;
  evidence_ref?: string;
};

export type TileEvent = {
  id: string;
  label: string;
  timestamp: string;
  kind: string;
  result?: string;
  requirement_id?: string;
  value?: number;
  severity?: string;
  evidence_ref?: string;
};

export type GraphTile = {
  schema_version: number | string;
  id: string;
  tile_id?: string;
  kind?: string;
  title?: string;
  renderer?: string;
  generated_at?: string;
  campaign_id?: string;
  card_id: string;
  level: string;
  t0: string;
  t1: string;
  time_window_ms?: number;
  axes?: unknown[];
  series: TileSeries[];
  bands: TileBand[];
  markers: GraphMarker[];
  events: TileEvent[];
  diagnostics: TileDiagnostics;
  provenance: TileProvenance;
};

// Wall and assignment types

export type WallConfig = {
  wall_id: string;
  label: string;
  preset?: boolean;
};

export type Assignment = {
  wall_id: string;
  tile_id: string;
  target_id: string;
  kind: string;
  param_id: number;
  device_id: string;
  instance: number;
  options: Record<string, unknown>;
};

// Signal catalogue types (device-agnostic)

export type SemanticSignal = {
  id: number;
  sid?: string;
  name: string;
  group: string;
  subgroup: string;
  role: "monitor" | "control";
  kind: "float" | "int" | "enum";
  unit?: string;
  type?: string;
  writable?: boolean;
  dangerous?: boolean;
  min?: number;
  max?: number;
  enum?: Record<string, string>;
  applicableModes?: string[];
  cmd?: string;
};

export type Channel = {
  device_id: string;
  instance: number;
  role: string;
  label: string;
  endpoint?: string;
};

// Adapter interfaces — consumers implement these

export interface SignalCatalogueAdapter {
  list(): SemanticSignal[];
  channels(): Channel[];
  channelsForSignal(signal: SemanticSignal): Channel[];
  subscribeLive(
    deviceId: string,
    paramId: number,
    instance: number,
    cb: (snap: { value: number | null; quality: string }) => void
  ): () => void;
  formatValue(value: number | null | undefined, unit?: string, paramId?: number): string;
  write?(deviceId: string, command: unknown, leaseToken: string): Promise<void>;
  roleForParam(paramId: number): string;
  colorForRole(role: string): string;
}

export interface TileAdapter {
  fetchTile(wallId: string, cardId: string, level: "live" | "minute" | "hour" | "three_hour" | "day" | "three_day"): Promise<GraphTile>;
}

export interface AssignmentsStoreOptions {
  namespace: string;
}
