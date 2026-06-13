import type { GraphTile, TilePoint, TileSeries } from "../types";

export const CANONICAL_TILE_RENDERER = "signalforge.tile.uplot";

export type SeriesRoleMeta = {
  readonly label: string;
  readonly rank: number;
  readonly className: string;
  readonly dash: string;
  readonly width: number;
  readonly opacity: number;
};

export const SERIES_ROLE_META: Readonly<Record<string, SeriesRoleMeta>> = Object.freeze({
  cmd: { label: "target / command", rank: 10, className: "cmd", dash: "6,4", width: 2.2, opacity: 0.98 },
  command: { label: "target / command", rank: 10, className: "cmd", dash: "6,4", width: 2.2, opacity: 0.98 },
  actual: { label: "actual", rank: 20, className: "actual", dash: "", width: 2.2, opacity: 0.98 },
  ghost: { label: "reference / sink", rank: 30, className: "ghost", dash: "2,4", width: 1.8, opacity: 0.86 },
  dut: { label: "power / load", rank: 40, className: "dut", dash: "", width: 2.0, opacity: 0.94 },
  aux: { label: "auxiliary", rank: 50, className: "aux", dash: "8,4", width: 1.8, opacity: 0.86 },
});

export type RenderedTileSeries = {
  key: string;
  tileId?: string;
  targetId?: unknown;
  label?: unknown;
  fullLabel?: unknown;
  role: string;
  seriesRole: string;
  roleRank: number;
  color?: string;
  unit: string;
  provenance?: unknown;
  source: unknown;
  paramId?: unknown;
  deviceId?: unknown;
  instance?: unknown;
  signalId?: unknown;
  history: {
    ts: number[];
    v: number[];
    q: string[];
  };
};

export type MeasuredElementSize = {
  width: number;
  height: number;
};

type LooseRecord = Record<string, unknown>;
type LooseTileSeries = TileSeries & LooseRecord;

export function graphSeriesIdentityKey(series: unknown): string {
  const item = recordOrEmpty(series);
  const source = recordOrEmpty(item.source_obj ?? item.source_ref ?? item.sourceRef ?? item.source);
  const sourceDevice = source.device_id ?? source.deviceId ?? item.device_id ?? item.deviceId;
  const sourceParam = source.param_id ?? source.paramId ?? item.param_id ?? item.paramId;
  const sourceInstance = source.instance ?? item.instance ?? 1;
  const deviceID = identityPart(sourceDevice);
  const paramID = identityPart(sourceParam);
  const instance = identityPart(sourceInstance, "1");
  if (deviceID && paramID) {
    return `${deviceID}:${paramID}:${instance}`;
  }

  const target = parseTargetId(item.target_id ?? item.targetId);
  if (target) return `${target.device_id}:${target.param_id}:${target.instance}`;

  const colon = parseColonKey(item.series_id ?? item.id ?? item.key);
  if (colon) return `${colon.device_id}:${colon.param_id}:${colon.instance}`;

  return text(item.series_id ?? item.id ?? item.key ?? item.target_id ?? item.targetId);
}

export function seriesRoleMeta(role?: string): SeriesRoleMeta {
  const key = role || "actual";
  const meta = SERIES_ROLE_META[key];
  return meta || SERIES_ROLE_META.actual;
}

export function seriesRoleColor(role?: string, fallback = "var(--series-actual)") {
  const meta = seriesRoleMeta(role);
  const css = meta.className ? `--series-${meta.className}` : "";
  if (!css || typeof document === "undefined") return fallback;
  return getComputedStyle(document.documentElement).getPropertyValue(css).trim() || fallback;
}

export function measuredElementSize(el: Element | null | undefined): MeasuredElementSize {
  if (!el) return { width: 0, height: 0 };
  const rect = typeof el.getBoundingClientRect === "function" ? el.getBoundingClientRect() : null;
  const html = el as HTMLElement;
  return {
    width: Math.floor((rect && rect.width) || html.clientWidth || html.offsetWidth || 0),
    height: Math.floor((rect && rect.height) || html.clientHeight || html.offsetHeight || 0),
  };
}

export function measuredElementWidth(el: Element | null | undefined) {
  return measuredElementSize(el).width;
}

export function emptyGraphTile(opts: LooseRecord = {}): GraphTile {
  const id = text(opts.tile_id ?? opts.tileId, "empty");
  const nowMs = Date.now();
  const timeWindowMs = numberOrUndefined(opts.time_window_ms ?? opts.timeWindowMs) ?? 90_000;
  const now = new Date(nowMs).toISOString();
  return {
    schema_version: "signalforge.graph_tile.v1",
    id,
    card_id: id,
    level: "live",
    t0: new Date(nowMs - timeWindowMs).toISOString(),
    t1: now,
    generated_at: now,
    renderer: CANONICAL_TILE_RENDERER,
    kind: "timeseries",
    tile_id: id,
    title: text(opts.title),
    time_window_ms: timeWindowMs,
    axes: Array.isArray(opts.axes) ? opts.axes : [],
    bands: [],
    markers: [],
    events: [],
    diagnostics: {
      status: "empty",
      point_count: 0,
      raw_point_count: 0,
      decimation: "none",
      freshness_ms: 0,
      renderer: CANONICAL_TILE_RENDERER,
      series_count: 0,
    },
    provenance: { source: "empty-graph-tile", generated_at: now },
    series: [],
  };
}

export type RetainGraphTileOptions = {
  readonly inViewport?: boolean;
};

export function hasRenderableGraphTile(tile: unknown): boolean {
  if (!isRecord(tile)) return false;
  if (Array.isArray(tile.series) && tile.series.length > 0) return true;
  if (Array.isArray(tile.bands) && tile.bands.length > 0) return true;
  if (Array.isArray(tile.markers) && tile.markers.length > 0) return true;
  if (Array.isArray(tile.events) && tile.events.length > 0) return true;
  const diagnostics = recordOrEmpty(tile.diagnostics);
  const seriesCount = numberOrUndefined(diagnostics.series_count ?? diagnostics.seriesCount);
  return Boolean(seriesCount && seriesCount > 0);
}

export function retainGraphTile<TileLike>(
  nextTile: TileLike,
  previousTile: TileLike | null | undefined,
  opts: RetainGraphTileOptions = {},
): TileLike | null | undefined {
  if (hasRenderableGraphTile(nextTile)) return nextTile;
  if (opts.inViewport !== false && hasRenderableGraphTile(previousTile)) return previousTile;
  return nextTile;
}

export function renderSeriesFromGraphTile(tile: unknown): RenderedTileSeries[] {
  if (!isRecord(tile) || !Array.isArray(tile.series)) return [];
  const normalizedTile = normalizeGraphTile(tile);
  return normalizedTile.series.map((series) => {
    const loose = series as LooseTileSeries;
    const legacyRole = text(loose.seriesRole ?? (series.role === "command" ? "cmd" : series.role), "actual");
    const source = recordOrEmpty(loose.source_obj);
    return {
      key: graphSeriesIdentityKey(series) || "series",
      tileId: normalizedTile.tile_id || normalizedTile.id,
      targetId: loose.target_id ?? loose.targetId,
      label: series.label,
      fullLabel: loose.full_label ?? loose.fullLabel ?? series.label,
      role: legacyRole,
      seriesRole: legacyRole,
      roleRank: numberOrUndefined(loose.role_rank) ?? seriesRoleMeta(legacyRole).rank,
      color: series.color || seriesRoleColor(legacyRole),
      unit: text(series.unit ?? series.units, "_"),
      provenance: loose.provenance || "",
      source: loose.source_obj ?? series.source ?? null,
      paramId: loose.param_id ?? source.param_id,
      deviceId: loose.device_id ?? source.device_id,
      instance: loose.instance ?? source.instance,
      signalId: loose.signal_id ?? source.signal_id,
      history: historyFromSeries(series),
    };
  }).sort((a, b) => {
    if (a.roleRank !== b.roleRank) return a.roleRank - b.roleRank;
    return String(a.tileId || a.key || a.label || "").localeCompare(String(b.tileId || b.key || b.label || ""));
  });
}

export function normalizeGraphTile(tile: unknown, opts: LooseRecord = {}): GraphTile {
  const fallback = emptyGraphTile({
    tile_id: opts.tile_id ?? opts.tileId,
    timeWindowMs: opts.timeWindowMs ?? opts.time_window_ms,
  });
  const sourceTile = recordOrEmpty(tile);
  const diagnostics = recordOrEmpty(sourceTile.diagnostics);
  const timeWindowMs = numberOrUndefined(sourceTile.time_window_ms ?? opts.timeWindowMs ?? opts.time_window_ms)
    ?? fallback.time_window_ms
    ?? 90_000;
  const series = (Array.isArray(sourceTile.series) ? sourceTile.series : [])
    .map(normalizeSeries)
    .filter((item) => (item.points?.length ?? 0) > 0 || (item.spans?.length ?? 0) > 0);
  const bands = Array.isArray(sourceTile.bands) ? sourceTile.bands as GraphTile["bands"] : [];
  const markers = Array.isArray(sourceTile.markers) ? sourceTile.markers as GraphTile["markers"] : [];
  const events = Array.isArray(sourceTile.events) ? sourceTile.events as GraphTile["events"] : [];
  const extents = [
    ...series.flatMap((item) => [
      ...(item.points || []).map((point) => point.timestamp),
      ...(item.spans || []).flatMap((span) => [span.start, span.end]),
    ]),
    ...bands.flatMap((band) => [band.start, band.end]),
    ...markers.map((marker) => marker.timestamp),
    ...events.map((event) => event.timestamp),
  ]
    .map((timestamp) => Date.parse(timestamp))
    .filter(Number.isFinite);
  const nowMs = Date.now();
  const parsedT0 = parseTime(sourceTile.t0);
  const parsedT1 = parseTime(sourceTile.t1);
  const t0Ms = parsedT0 ?? (extents.length ? Math.min(...extents) : nowMs - timeWindowMs);
  const t1Ms = parsedT1 ?? (extents.length ? Math.max(...extents) : nowMs);
  const t0 = new Date(t0Ms).toISOString();
  const t1 = new Date(Math.max(t1Ms, t0Ms + 1)).toISOString();
  const pointCount = series.reduce((acc, item) => acc + (item.points?.length || 0), 0);
  const diagnosticsStatus = text(diagnostics.status, series.length > 0 ? "ok" : "empty");
  return {
    ...fallback,
    ...(sourceTile as Partial<GraphTile>),
    schema_version: schemaVersion(sourceTile.schema_version, fallback.schema_version),
    id: text(sourceTile.id ?? sourceTile.tile_id, fallback.id),
    card_id: text(sourceTile.card_id ?? sourceTile.tile_id ?? sourceTile.id, fallback.card_id),
    level: text(sourceTile.level, "live"),
    t0,
    t1,
    generated_at: text(sourceTile.generated_at, new Date(nowMs).toISOString()),
    renderer: CANONICAL_TILE_RENDERER,
    kind: text(sourceTile.kind, "timeseries"),
    tile_id: text(sourceTile.tile_id ?? sourceTile.id, fallback.tile_id),
    title: text(sourceTile.title, fallback.title),
    time_window_ms: timeWindowMs,
    axes: Array.isArray(sourceTile.axes) ? sourceTile.axes : fallback.axes,
    bands,
    markers,
    events,
    diagnostics: {
      ...fallback.diagnostics,
      ...diagnostics,
      status: diagnosticsStatus,
      point_count: pointCount,
      raw_point_count: numberOrUndefined(diagnostics.raw_point_count) ?? pointCount,
      decimation: text(diagnostics.decimation, "none"),
      renderer: CANONICAL_TILE_RENDERER,
      series_count: series.length,
    },
    provenance: isRecord(sourceTile.provenance) ? sourceTile.provenance as GraphTile["provenance"] : fallback.provenance,
    series,
  };
}

function normalizeSeries(value: unknown): LooseTileSeries {
  const series = recordOrEmpty(value);
  const sourceObj = recordFrom(series.source_obj) ?? recordFrom(series.source_ref) ?? recordFrom(series.source);
  const legacyRole = text(series.role ?? series.seriesRole, "actual");
  const role = canonicalRole(legacyRole);
  const id = text(series.series_id ?? series.id ?? series.key ?? series.target_id ?? series.targetId ?? series.label, "series");
  const unit = text(series.unit ?? series.units, "_");
  const points = normalizePoints(series);
  return {
    ...series,
    id,
    series_id: series.series_id || id,
    label: text(series.label, id),
    role,
    seriesRole: legacyRole,
    unit,
    units: unit,
    axis_id: text(series.axis_id ?? axisIdForSeries({ ...series, id, role, seriesRole: legacyRole }, unit)),
    source: stringifySource(series.source_ref ?? series.source ?? sourceObj ?? id),
    source_obj: sourceObj,
    color: typeof series.color === "string" ? series.color : undefined,
    points,
    spans: Array.isArray(series.spans) ? series.spans as TileSeries["spans"] : [],
  } as unknown as LooseTileSeries;
}

function canonicalRole(role: unknown) {
  const value = text(role, "actual");
  if (value === "cmd") return "command";
  return value || "actual";
}

function axisIdForSeries(series: LooseRecord, unit: unknown) {
  const explicit = series.axis_id ?? series.axisId;
  if (explicit) return text(explicit);
  const u = text(unit ?? series.unit ?? series.units).trim().toLowerCase();
  const label = [
    series.id,
    series.series_id,
    series.key,
    series.target_id,
    series.targetId,
    series.label,
    series.full_label,
    series.fullLabel,
  ].filter(Boolean).join(" ").toLowerCase();
  if (series.role === "counter" || series.seriesRole === "counter" || series.kind === "counter") return "counter";
  if (u === "a" || u === "amp" || u === "amps") return "current_a";
  if (u === "v" || u === "volt" || u === "volts") return "voltage_v";
  if (u === "w" || u === "watt" || u === "watts") return "power_w";
  if (u === "%" || u === "percent") return "percent";
  if (u === "ms" || u === "millisecond" || u === "milliseconds") return "bus_ms";
  if (u === "s" || u === "sec" || u === "secs" || u === "second" || u === "seconds") return "seconds";
  if (u === "mbar" || u === "millibar" || u === "millibars") return "pressure_log";
  if (u === "mbar/min" || u === "mbar/minute" || u === "millibar/min" || u === "millibars/minute") return "pressure_rate_log";
  if (u === "bar") return "pressure_bar";
  if (u.includes("deg") || u === "c" || u === "degc" || u === "deg c" || u === "°c" || u === "° c") return "temperature_c";
  if (label.includes("counter")) return "counter";
  if (label.includes("pressure") || label.includes("vacuum")) return "pressure_log";
  return "generic_numeric";
}

function normalizePoints(series: LooseRecord): TilePoint[] {
  if (Array.isArray(series.points) && series.points.length) {
    return series.points.flatMap((point) => normalizePoint(point));
  }
  const history = recordOrEmpty(series.history);
  const ts = Array.isArray(history.ts) ? history.ts : [];
  const values = Array.isArray(history.v) ? history.v : [];
  return values.flatMap((value, idx) => normalizePoint({ t: ts[idx], v: value }));
}

function normalizePoint(point: unknown): TilePoint[] {
  const item = recordOrEmpty(point);
  const rawTimestamp = item.timestamp ?? item.t ?? item.time;
  const rawValue = item.value ?? item.v ?? item.y;
  if (rawTimestamp === undefined || rawTimestamp === null || rawTimestamp === "") return [];
  if (rawValue === undefined || rawValue === null || rawValue === "") return [];
  const value = Number(rawValue);
  const timeMs = typeof rawTimestamp === "number" ? rawTimestamp : Date.parse(String(rawTimestamp || ""));
  if (!Number.isFinite(value) || !Number.isFinite(timeMs)) return [];
  return [{ timestamp: new Date(timeMs).toISOString(), value }];
}

function parseTargetId(value: unknown) {
  const match = String(value || "").match(/^([^@]+)@([^/]+)\/([^/]+)$/);
  if (!match) return null;
  const paramID = identityPart(match[1]);
  const deviceID = identityPart(match[2]);
  const instance = identityPart(match[3], "1");
  if (!deviceID || !paramID) return null;
  return { param_id: paramID, device_id: deviceID, instance };
}

function parseColonKey(value: unknown) {
  const parts = String(value || "").split(":");
  if (parts.length < 3) return null;
  const instance = identityPart(parts[parts.length - 1], "1");
  const paramID = identityPart(parts[parts.length - 2]);
  const deviceID = identityPart(parts.slice(0, -2).join(":"));
  if (!deviceID || !paramID) return null;
  return { device_id: deviceID, param_id: paramID, instance };
}

function identityPart(value: unknown, fallback = "") {
  if (value === undefined || value === null || value === "") return fallback;
  return String(value).trim() || fallback;
}

function historyFromSeries(series: TileSeries) {
  const points = series.points || [];
  return {
    ts: points.map((point) => Date.parse(point.timestamp)),
    v: points.map((point) => point.value),
    q: points.map(() => "ok"),
  };
}

function stringifySource(source: unknown): string {
  if (!source) return "";
  if (typeof source === "string") return source;
  if (!isRecord(source)) return String(source);
  const device = source.device_id || source.deviceId || "";
  const param = source.param_id || source.paramId || "";
  const instance = source.instance || "";
  const endpoint = source.endpoint || "";
  const signal = source.signal_id || source.signalId || "";
  return `device=${device} param=${param} instance=${instance} signal=${signal} endpoint=${endpoint}`.trim();
}

function parseTime(value: unknown) {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value !== "string" || !value) return undefined;
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function text(value: unknown, fallback = "") {
  if (value === undefined || value === null || value === "") return fallback;
  return String(value);
}

function numberOrUndefined(value: unknown) {
  if (value === undefined || value === null || value === "") return undefined;
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : undefined;
}

function schemaVersion(value: unknown, fallback: GraphTile["schema_version"]): GraphTile["schema_version"] {
  return typeof value === "string" || typeof value === "number" ? value : fallback;
}

function recordOrEmpty(value: unknown): LooseRecord {
  return isRecord(value) ? value : {};
}

function recordFrom(value: unknown): LooseRecord | undefined {
  return isRecord(value) ? value : undefined;
}

function isRecord(value: unknown): value is LooseRecord {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
