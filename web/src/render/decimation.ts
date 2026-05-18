import type { GraphTile, TileSeries } from "../types";

type TilePoint = NonNullable<TileSeries["points"]>[number];
type TilePointRun = NonNullable<TileSeries["points"]>;
export type PreparedSeriesPoint = {
  t: number;
  v: number;
};
type GapAwareSegment =
  | { kind: "run"; points: TilePointRun }
  | { kind: "gap"; point: TilePoint };

export function viewportSeries(tile: GraphTile, series: TileSeries, viewportWidth: number): TileSeries {
  const points = series.points ?? [];
  if (points.length < 4 || isDiscreteSeries(series)) return series;
  const budget = Math.max(180, Math.min(points.length, Math.round(viewportWidth * 1.65)));
  if (points.length <= budget) return series;
  const yValue = (value: number) => decimationValue(tile, series, value);
  if (isPressureLogAxis(series.axis_id)) {
    return { ...series, points: lttbPreservingGaps(points, budget, yValue) };
  }
  return { ...series, points: lttb(points, budget, yValue) };
}

export function lttb(points: TileSeries["points"], threshold: number, yValue: (value: number) => number): TileSeries["points"] {
  if (!points || threshold >= points.length || threshold < 3) return points;
  const parsed = points
    .map((point) => ({ point, x: Date.parse(point.timestamp), y: yValue(point.value) }))
    .filter((point) => Number.isFinite(point.x) && Number.isFinite(point.y));
  if (parsed.length <= threshold) return parsed.map((item) => item.point);
  const sampled = [parsed[0].point];
  const bucketSize = (parsed.length - 2) / (threshold - 2);
  let a = 0;
  for (let i = 0; i < threshold - 2; i++) {
    const rangeStart = Math.floor((i + 0) * bucketSize) + 1;
    const rangeEnd = Math.floor((i + 1) * bucketSize) + 1;
    const nextRangeStart = Math.floor((i + 1) * bucketSize) + 1;
    const nextRangeEnd = Math.floor((i + 2) * bucketSize) + 1;
    const range = parsed.slice(rangeStart, Math.min(rangeEnd, parsed.length - 1));
    const nextRange = parsed.slice(nextRangeStart, Math.min(nextRangeEnd, parsed.length));
    const avgX = nextRange.reduce((sum, point) => sum + point.x, 0) / Math.max(1, nextRange.length);
    const avgY = nextRange.reduce((sum, point) => sum + point.y, 0) / Math.max(1, nextRange.length);
    const anchor = parsed[a];
    let selected = range[0] ?? parsed[Math.min(rangeStart, parsed.length - 2)];
    let selectedIndex = range.length ? rangeStart : Math.min(rangeStart, parsed.length - 2);
    let maxArea = -1;
    range.forEach((candidate, offset) => {
      const area = Math.abs((anchor.x - avgX) * (candidate.y - anchor.y) - (anchor.x - candidate.x) * (avgY - anchor.y));
      if (area > maxArea) {
        maxArea = area;
        selected = candidate;
        selectedIndex = rangeStart + offset;
      }
    });
    sampled.push(selected.point);
    a = selectedIndex;
  }
  sampled.push(parsed[parsed.length - 1].point);
  return sampled;
}

export function lttbPreservingGaps(points: TileSeries["points"], threshold: number, yValue: (value: number) => number): TileSeries["points"] {
  if (!points || threshold >= points.length || threshold < 3) return points;

  const segments: GapAwareSegment[] = [];
  let currentRun: TilePointRun = [];
  const flushRun = () => {
    if (currentRun.length) {
      segments.push({ kind: "run", points: currentRun });
      currentRun = [];
    }
  };

  for (const point of points) {
    const x = Date.parse(point.timestamp);
    const y = yValue(point.value);
    if (Number.isFinite(x) && Number.isFinite(y)) {
      currentRun.push(point);
      continue;
    }
    flushRun();
    if (Number.isFinite(x)) segments.push({ kind: "gap", point });
  }
  flushRun();

  const runs = segments.filter((segment): segment is { kind: "run"; points: TilePointRun } => segment.kind === "run");
  const gapCount = segments.length - runs.length;
  if (!gapCount) return lttb(points, threshold, yValue);

  const finiteCount = runs.reduce((sum, run) => sum + run.points.length, 0);
  const available = Math.max(0, threshold - gapCount);
  const budgets = runs.map((run) => {
    if (run.points.length <= 2) return run.points.length;
    const proportional = finiteCount > 0 ? Math.round((run.points.length / finiteCount) * available) : run.points.length;
    return Math.min(run.points.length, Math.max(3, proportional));
  });

  const minBudget = (run: { points: TilePointRun }) => run.points.length <= 2 ? run.points.length : 3;
  let total = gapCount + budgets.reduce((sum, value) => sum + value, 0);
  while (total > threshold) {
    let candidate = -1;
    for (let index = 0; index < budgets.length; index += 1) {
      if (budgets[index] <= minBudget(runs[index])) continue;
      if (candidate === -1 || budgets[index] > budgets[candidate]) candidate = index;
    }
    if (candidate === -1) break;
    budgets[candidate] -= 1;
    total -= 1;
  }
  if (total > threshold) {
    return tightSamplePreservingRepresentativeGaps(segments, threshold, yValue);
  }

  const sampled: TilePointRun = [];
  let runIndex = 0;
  for (const segment of segments) {
    if (segment.kind === "gap") {
      sampled.push(segment.point);
      continue;
    }
    const budget = budgets[runIndex] ?? segment.points.length;
    sampled.push(...(budget >= segment.points.length ? segment.points : lttb(segment.points, budget, yValue) ?? []));
    runIndex += 1;
  }
  return sampled;
}

function tightSamplePreservingRepresentativeGaps(segments: GapAwareSegment[], threshold: number, yValue: (value: number) => number): TileSeries["points"] {
  const runs = segments.filter((segment): segment is { kind: "run"; points: TilePointRun } => segment.kind === "run");
  const gaps = segments
    .map((segment, index) => ({ segment, index }))
    .filter((item): item is { segment: { kind: "gap"; point: TilePoint }; index: number } => item.segment.kind === "gap");
  if (!runs.length) return gaps.slice(0, threshold).map((item) => item.segment.point);

  const finiteCount = runs.reduce((sum, run) => sum + run.points.length, 0);
  const maxGapBudget = Math.max(0, threshold - 1);
  const proportionalGapBudget = Math.round(threshold * (gaps.length / (gaps.length + runs.length)));
  const gapBudget = Math.min(gaps.length, maxGapBudget, Math.max(1, proportionalGapBudget));
  const keptGapIndexes = new Set<number>();
  if (gapBudget > 0) {
    for (let i = 0; i < gapBudget; i += 1) {
      const gap = gaps[Math.floor((i * gaps.length) / gapBudget)];
      if (gap) keptGapIndexes.add(gap.index);
    }
  }

  const runBudgets = allocateRunBudgets(runs, Math.max(0, threshold - keptGapIndexes.size), finiteCount);
  const sampled: TilePointRun = [];
  let runIndex = 0;
  for (let index = 0; index < segments.length && sampled.length < threshold; index += 1) {
    const segment = segments[index];
    if (segment.kind === "gap") {
      if (keptGapIndexes.has(index)) sampled.push(segment.point);
      continue;
    }
    const budget = Math.min(runBudgets[runIndex] ?? 0, threshold - sampled.length);
    sampled.push(...sampleRun(segment.points, budget, yValue));
    runIndex += 1;
  }
  return sampled.slice(0, threshold);
}

function allocateRunBudgets(runs: Array<{ points: TilePointRun }>, budget: number, finiteCount: number) {
  const budgets = runs.map(() => 0);
  if (budget <= 0) return budgets;
  if (budget < runs.length) {
    for (let i = 0; i < budget; i += 1) {
      const index = Math.floor((i * runs.length) / budget);
      budgets[index] = 1;
    }
    return budgets;
  }
  runs.forEach((run, index) => {
    budgets[index] = Math.min(run.points.length, 1);
  });
  let remaining = budget - budgets.reduce((sum, value) => sum + value, 0);
  while (remaining > 0) {
    let candidate = -1;
    let bestNeed = 0;
    runs.forEach((run, index) => {
      const proportional = finiteCount > 0 ? (run.points.length / finiteCount) * budget : budget / Math.max(1, runs.length);
      const need = Math.min(run.points.length, Math.max(1, Math.round(proportional))) - budgets[index];
      if (need > bestNeed) {
        bestNeed = need;
        candidate = index;
      }
    });
    if (candidate === -1) break;
    budgets[candidate] += 1;
    remaining -= 1;
  }
  return budgets;
}

function sampleRun(points: TilePointRun, budget: number, yValue: (value: number) => number): TilePointRun {
  if (budget <= 0) return [];
  if (budget >= points.length) return points;
  if (budget === 1) return [points[Math.floor((points.length - 1) / 2)]];
  if (budget === 2) return [points[0], points[points.length - 1]];
  return lttb(points, budget, yValue) ?? [];
}

export function decimationValue(_tile: GraphTile, series: TileSeries, value: number) {
  if (isPressureLogAxis(series.axis_id)) return value > 0 ? Math.log10(value) : Number.NaN;
  return value;
}

export function resampleSeries(tile: GraphTile, series: TileSeries, xValues: number[], currentTimeMs?: number): Array<number | null> {
  return resamplePreparedSeries(tile, series, prepareSeriesPoints(series), xValues, currentTimeMs);
}

export function prepareSeriesPoints(series: TileSeries): PreparedSeriesPoint[] {
  return [...(series.points ?? [])]
    .map((point) => ({ t: Date.parse(point.timestamp), v: interpolationValue(series, point.value) }))
    .filter((point) => Number.isFinite(point.t))
    .sort((a, b) => a.t - b.t);
}

export function resamplePreparedSeries(tile: GraphTile, series: TileSeries, points: PreparedSeriesPoint[], xValues: number[], currentTimeMs?: number): Array<number | null> {
  if (!points.length) return xValues.map(() => null);

  const stepped = isDiscreteSeries(series) || series.render_kind === "swimlane";
  const isFutureVisible = series.role === "ghost" || commandCenterProjectedSeries(tile, series);
  const gapThreshold = commandCenterTraceGapMs(tile, series);
  let cursor = 0;
  return xValues.map((x) => {
    if (Number.isFinite(currentTimeMs) && x > (currentTimeMs as number) && !isFutureVisible) return null;
    while (cursor + 1 < points.length && points[cursor + 1].t <= x) cursor += 1;
    const current = points[cursor];
    const next = points[Math.min(cursor + 1, points.length - 1)];
    if (x < points[0].t || x > points[points.length - 1].t) return null;
    if (gapThreshold > 0 && next.t - current.t > gapThreshold && x > current.t && x < next.t) return null;
    if (x === current.t) return Number.isFinite(current.v) ? valueFromInterpolation(series, current.v) : null;
    if (!Number.isFinite(current.v) || !Number.isFinite(next.v)) return null;
    if (stepped || next.t === current.t) return valueFromInterpolation(series, current.v);
    const ratio = (x - current.t) / (next.t - current.t);
    const interpolated = current.v + (next.v - current.v) * Math.max(0, Math.min(1, ratio));
    return valueFromInterpolation(series, interpolated);
  });
}

export function commandCenterGapBreaks(tile: GraphTile, series: TileSeries) {
  return commandCenterGapBreaksFromPoints(tile, series, prepareSeriesPoints(series));
}

export function commandCenterGapBreaksFromPoints(tile: GraphTile, series: TileSeries, points: PreparedSeriesPoint[]) {
  const gapThreshold = commandCenterTraceGapMs(tile, series);
  if (gapThreshold <= 0) return [];
  const breaks: number[] = [];
  for (let i = 1; i < points.length; i += 1) {
    if (points[i].t - points[i - 1].t > gapThreshold) {
      breaks.push(points[i - 1].t + 1, points[i].t - 1);
    }
  }
  return breaks;
}

export function commandCenterTraceGapMs(tile: GraphTile, series: TileSeries) {
  if (tile.campaign_id !== "command_center_fat") return 0;
  if (series.render_kind === "swimlane" || isDiscreteSeries(series) || series.role === "event") return 0;
  return 2 * 60 * 60 * 1000;
}

export function commandCenterProjectedSeries(tile: GraphTile, series: TileSeries) {
  return tile.campaign_id === "command_center_fat" && series.role === "command";
}

// displayValue is used internally by resampleSeries; also re-exported for uPlotAdapter
export function displayValue(_tile: GraphTile, series: TileSeries, value: number) {
  if (isPressureLogAxis(series.axis_id)) return value > 0 ? value : Number.NaN;
  return value;
}

export function isDiscreteSeries(series: TileSeries) {
  return Boolean(series.step) || series.render_kind === "counter" || series.kind === "counter" || series.role === "counter";
}

export function interpolationValue(series: TileSeries, value: number) {
  if (!isPressureLogAxis(series.axis_id)) return value;
  return value > 0 ? Math.log10(value) : Number.NaN;
}

export function valueFromInterpolation(series: TileSeries, value: number) {
  if (!Number.isFinite(value)) return Number.NaN;
  if (!isPressureLogAxis(series.axis_id)) return value;
  return 10 ** value;
}

export function isPressureAxis(axisID?: string) {
  return axisID === "pressure_mbar"
    || axisID === "pressure_rate"
    || axisID === "pressure_log"
    || axisID === "pressure_rate_log";
}

export function isPressureLogAxis(axisID?: string) {
  return axisID === "pressure_log" || axisID === "pressure_rate_log";
}
