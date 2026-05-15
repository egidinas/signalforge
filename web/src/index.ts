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
  // Signal catalogue
  SemanticSignal, Channel,
  // Adapter interfaces
  SignalCatalogueAdapter, TileAdapter, AssignmentsStoreOptions,
} from "./types";

// Hooks
export { loadAssignments, saveAssignments, makeAssignment, useAssignments } from "./dict/useAssignments";
export type { AssignmentsHandle } from "./dict/useAssignments";
export { loadWalls, saveWalls, useWalls } from "./walls/useWalls";
export type { WallsHandle } from "./walls/useWalls";
export { useTileSeries } from "./tiles/useTileSeries";
export type { TileState } from "./tiles/useTileSeries";

// Tile client
export { TileClient, pickTileLevel } from "./tiles/TileClient";
export type { TileLevel } from "./tiles/TileClient";

// Render utilities
export { uplotData, drawTileOverlays } from "./render/uPlotAdapter";

// React components
export { SignalDictionary } from "./dict/SignalDictionary";
export type { SignalDictionaryProps } from "./dict/SignalDictionary";
export { WallManager } from "./walls/WallManager";
export type { WallManagerProps } from "./walls/WallManager";
export { UPlotTileRenderer } from "./render/UPlotTileRenderer";
export type { UPlotTileRendererProps } from "./render/UPlotTileRenderer";
