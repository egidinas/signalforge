import { decimationValue, lttb, lttbPreservingGaps, resampleSeries, viewportSeries } from "../src/render/decimation";
import type { GraphTile, TileSeries } from "../src/types";

describe("decimation", () => {
  const tile = (series: TileSeries[], campaign_id = "generic"): GraphTile => ({
    schema_version: "signalforge.graph_tile.v1",
    id: "tile",
    card_id: "tile",
    campaign_id,
    level: "live",
    t0: "2024-01-01T00:00:00.000Z",
    t1: "2024-01-01T04:00:00.000Z",
    renderer: "signalforge.tile.uplot",
    series,
    bands: [],
    markers: [],
    events: [],
    diagnostics: {},
    provenance: {},
  } as GraphTile);

  it("preserves LTTB output shape after indexed bucket selection", () => {
    const points: NonNullable<TileSeries["points"]> = Array.from({ length: 200 }, (_, index) => ({
      timestamp: new Date(1_700_000_000_000 + index * 1000).toISOString(),
      value: Math.sin(index / 5) * 10 + index / 20,
    }));

    const decimated = lttb(points, 24, (value) => value) ?? [];

    expect(decimated).toHaveLength(24);
    expect(decimated[0]).toBe(points[0]);
    expect(decimated[decimated.length - 1]).toBe(points[points.length - 1]);
    expect(new Set(decimated.map((point) => point.timestamp)).size).toBe(decimated.length);
  });

  it("keeps generic LTTB output within budget after filtering invalid samples", () => {
    const baseTime = Date.parse("2024-01-01T00:00:00.000Z");
    const points: NonNullable<TileSeries["points"]> = Array.from({ length: 20 }, (_, index) => ({
      timestamp: index % 2 === 0 ? new Date(baseTime + index * 1000).toISOString() : "invalid",
      value: index < 12 ? Number.NaN : index,
    }));

    const decimated = lttb(points, 5, (value) => value) ?? [];

    expect(decimated.length).toBeLessThanOrEqual(5);
    expect(decimated.every((point) => Number.isFinite(Date.parse(point.timestamp)) && Number.isFinite(point.value))).toBe(true);
  });

  it("treats role-only counters as discrete data", () => {
    const points: NonNullable<TileSeries["points"]> = Array.from({ length: 250 }, (_, index) => ({
      timestamp: new Date(1_700_000_000_000 + index * 1000).toISOString(),
      value: index,
    }));
    const series: TileSeries = {
      id: "packets",
      label: "Packets",
      role: "counter",
      source: "fixture",
      axis_id: "counter",
      points,
    };
    const graphTile = tile([series]);

    expect(viewportSeries(graphTile, series, 20).points).toBe(points);
    expect(resampleSeries(graphTile, series, [Date.parse(points[2].timestamp) + 500])).toEqual([2]);
  });

  it("uses canonical pressure axes for log-space decimation", () => {
    expect(decimationValue(tile([]), { id: "p", label: "Pressure", role: "actual", source: "fixture", axis_id: "pressure_log", points: [] }, 100)).toBe(2);
    expect(decimationValue(tile([]), { id: "r", label: "Pressure rate", role: "actual", source: "fixture", axis_id: "pressure_rate_log", points: [] }, 10)).toBe(1);
  });

  it("interpolates pressure axes in log space and keeps invalid samples as gaps", () => {
    const points = [
      { timestamp: "2024-01-01T00:00:00.000Z", value: 1 },
      { timestamp: "2024-01-01T00:05:00.000Z", value: 10 },
      { timestamp: "2024-01-01T00:10:00.000Z", value: 0 },
      { timestamp: "2024-01-01T00:20:00.000Z", value: 100 },
    ];
    const series: TileSeries = {
      id: "pressure",
      label: "Pressure",
      role: "actual",
      source: "fixture",
      axis_id: "pressure_log",
      points,
    };
    const graphTile = tile([series]);

    expect(resampleSeries(graphTile, series, [Date.parse("2024-01-01T00:05:00.000Z")])).toEqual([10]);
    expect(resampleSeries(graphTile, series, [Date.parse("2024-01-01T00:10:00.000Z")])).toEqual([null]);
    expect(resampleSeries(graphTile, series, [Date.parse("2024-01-01T00:15:00.000Z")])).toEqual([null]);
    expect(resampleSeries(graphTile, series, [Date.parse("2024-01-01T00:20:00.000Z")])).toEqual([100]);

    const cleanSeries = { ...series, points: [points[0], points[3]] };
    const [midpoint] = resampleSeries(graphTile, cleanSeries, [Date.parse("2024-01-01T00:10:00.000Z")]);
    expect(midpoint).toBeCloseTo(10, 5);
  });

  it("preserves pressure gap sentinels through viewport decimation", () => {
    const baseTime = Date.parse("2024-01-01T00:00:00.000Z");
    const points: NonNullable<TileSeries["points"]> = Array.from({ length: 260 }, (_, index) => ({
      timestamp: new Date(baseTime + index * 1000).toISOString(),
      value: index === 130 ? 0 : 10 ** (-6 + (index / 259) * 8),
    }));
    const series: TileSeries = {
      id: "pressure",
      label: "Pressure",
      role: "actual",
      source: "fixture",
      axis_id: "pressure_log",
      points,
    };
    const graphTile = tile([series]);
    const decimated = viewportSeries(graphTile, series, 20);
    const gapTime = Date.parse(points[130].timestamp);

    expect(decimated.points?.some((point) => point.value === 0)).toBe(true);
    expect(resampleSeries(graphTile, decimated, [gapTime, gapTime + 500])).toEqual([null, null]);
  });

  it("keeps gap-preserving pressure decimation within the viewport budget under dense gaps", () => {
    const baseTime = Date.parse("2024-01-01T00:00:00.000Z");
    const points: NonNullable<TileSeries["points"]> = Array.from({ length: 40 }, (_, index) => ({
      timestamp: new Date(baseTime + index * 1000).toISOString(),
      value: index % 2 === 0 ? 10 ** (index / 10) : 0,
    }));

    const decimated = lttbPreservingGaps(points, 12, (value) => (value > 0 ? Math.log10(value) : Number.NaN)) ?? [];

    expect(decimated.length).toBeLessThanOrEqual(12);
    expect(decimated.some((point) => point.value === 0)).toBe(true);
    expect(decimated.some((point) => point.value > 0)).toBe(true);
  });
});
