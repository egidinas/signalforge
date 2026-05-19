// signalforge-web public API

// Types
export type {
  // Tile schema
  GraphTile, TileSeries, TilePoint, TileSpan, TileBand, TileEvent,
  TileDiagnostics, TileProvenance,
  // Graph model (render layer)
  HeroGraphModel, GraphYAxis, GraphTrace, GraphTimeAxis, GraphBand,
  GraphMarker, CompanionGraphGroup, GraphWallModel, GraphWallCard,
  GraphWallSignal, GraphSection, GraphTileCardRef,
  // Wall / assignment
  WallConfig, Assignment,
  // Public semantic contracts
  LifecycleState, LifecycleStep, WriteLifecycleContract,
  SemanticValueQuality, SemanticValueCounterpart, SemanticValueHoverMeta,
  RouteState, RouteCandidate, RouteRedundancyContract,
  JsonSignalCatalogue, JsonSignalCatalogueEntry,
  SourceCatalogue, SourceCatalogueCapabilities, SourceCatalogueEntry, SourceCatalogueTransportPath,
  SemanticOverlayBundle, SemanticOverlayEntry, SemanticOverlayTarget, SemanticOverlayStoreOptions,
  // Signal catalogue
  SemanticSignal, Channel,
  // Adapter interfaces
  SignalCatalogueAdapter, TileAdapter, AssignmentsStoreOptions,
} from "./types";

// Hooks
export { loadAssignments, saveAssignments, makeAssignment, useAssignments } from "./dict/useAssignments";
export type { AssignmentsHandle } from "./dict/useAssignments";
export { normalizeJsonSignalCatalogue } from "./catalogue/json";
export type { JsonSignalCatalogueInput } from "./catalogue/json";
export {
  CURRENT_SOURCE_PROJECTION_SCHEMA_VERSION,
  buildSignalProjectionTree,
  normalizeSignalProjectionBundle,
  signalProjectionPathKey,
  signalProjectionRefKey,
  validateSignalProjectionBundle,
} from "./catalogue/projection";
export type {
  SignalProjectionBundle,
  SignalProjectionBundleInput,
  SignalProjectionMapping,
  SignalProjectionMappingInput,
  SignalProjectionMappingKind,
  SignalProjectionPathSegment,
  SignalProjectionPathSegmentInput,
  SignalProjectionSignalID,
  SignalProjectionSignalRef,
  SignalProjectionSourceCatalogueEntryLike,
  SignalProjectionSourceCatalogueLike,
  SignalProjectionSourceFamily,
  SignalProjectionTreeNode,
  SignalProjectionValidationIssue,
  SignalProjectionValidationOptions,
  SignalProjectionValidationResult,
} from "./catalogue/projection";
export {
  semanticOverlayTargetKey,
  channelSemanticTarget,
  normalizeSemanticOverlayEntry,
  normalizeSemanticOverlayBundle,
  overlayEntryForTarget,
  upsertSemanticOverlayEntry,
  removeSemanticOverlayEntry,
  loadSemanticOverlay,
  saveSemanticOverlay,
  useSemanticOverlay,
} from "./catalogue/semanticOverlay";
export type { SemanticOverlayHandle } from "./catalogue/semanticOverlay";
export { loadWalls, saveWalls, useWalls } from "./walls/useWalls";
export type { WallsHandle } from "./walls/useWalls";
export { useTileSeries } from "./tiles/useTileSeries";
export type { TileState } from "./tiles/useTileSeries";

// Tile client
export { DEFAULT_TILE_LEVELS, TileClient, pickTileLevel } from "./tiles/TileClient";
export type { TileLevel, TileLevelSpec } from "./tiles/TileClient";

// Render utilities
export {
  uplotData,
  drawTileOverlays,
  seriesDrawOrder,
  lineWidthFor,
  sharedTimeGrid,
  buildScales,
  buildAxes,
  paddedRange,
  logScale,
  logSplits,
  ySplits,
  axisLabel,
  scaleForSeries,
  stateBlocks,
  inTimeRange,
  renderKindFor,
} from "./render/uPlotAdapter";
export {
  CANONICAL_TILE_RENDERER,
  SERIES_ROLE_META,
  seriesRoleMeta,
  seriesRoleColor,
  measuredElementSize,
  measuredElementWidth,
  graphSeriesIdentityKey,
  emptyGraphTile,
  renderSeriesFromGraphTile,
  normalizeGraphTile,
} from "./render/tileModel";
export type { SeriesRoleMeta, RenderedTileSeries, MeasuredElementSize } from "./render/tileModel";
export {
  clampRange,
  timeTicks,
  chooseTickStep,
  tickLabel,
  zoomRangeAroundAnchor,
  TimeAxisTrack,
  HeroTopTimeAxis,
  SharedTimeAxis,
} from "./render/timeAxis";
export {
  roleColors,
  signalColors,
  distinctivePalette,
  palette,
  paletteForID,
  colorForSignal,
  semanticColor,
  signalPriority,
  orderLegendSignals,
  graphCardPriority,
  graphSectionPriority,
  graphSectionRank,
  graphCardRank,
  cardPriority,
  tileCardPriority,
  eventColor,
  blockLabel,
} from "./render/visualPolicy";
export {
  viewportSeries,
  commandCenterGapBreaks,
  resampleSeries,
  decimationValue,
  commandCenterProjectedSeries,
  displayValue,
  lttb,
  commandCenterTraceGapMs,
  interpolationValue,
  isDiscreteSeries,
  valueFromInterpolation,
} from "./render/decimation";
export {
  markerColor,
  operatorMarkerLines,
  formatMarkerDateTime,
  placeMarkerLabel,
  rectanglesOverlap,
  fitCanvasText,
  shortGateLabel,
  legendReadouts,
  clampTime,
  rawValueAt,
  stateAt,
  stateLabel,
  formatLegendValue,
  formatScientific,
  formatPressure,
  unitForAxis,
} from "./render/markers";

// UI fundamentals
export { SF_UI_CLASSNAMES, SF_UI_TOKENS } from "./ui/fundamentals";
export { Sparkline } from "./ui/Sparkline";
export type { SparklineProps } from "./ui/Sparkline";
export { WriteLifecycleTrace } from "./ui/WriteLifecycleTrace";
export type { WriteLifecycleTraceProps, WriteLifecyclePhase } from "./ui/WriteLifecycleTrace";
export { HistoryController } from "./ui/HistoryController";
export type { HistoryControllerProps, HistoryRange, HistoryAvailability } from "./ui/HistoryController";
export { SemanticDiscoveryTree } from "./ui/SemanticDiscoveryTree";
export type { SemanticDiscoveryNode, SemanticDiscoveryTreeProps } from "./ui/SemanticDiscoveryTree";

// React components
export { SignalDictionary } from "./dict/SignalDictionary";
export type { SignalDictionaryProps } from "./dict/SignalDictionary";
export { WallManager } from "./walls/WallManager";
export type { WallManagerProps } from "./walls/WallManager";
export { UPlotTileRenderer } from "./render/UPlotTileRenderer";
export type { UPlotTileRendererProps } from "./render/UPlotTileRenderer";
