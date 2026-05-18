import type uPlot from "uplot";
import { readFileSync } from "node:fs";
import { drawTileOverlays, paddedRange, uplotData } from "../src/render/uPlotAdapter";
import { normalizeGraphTile } from "../src/render/tileModel";
import type { GraphTile } from "../src/types";

function fakeContext() {
  const fn = () => undefined;
  const fillStyles: string[] = [];
  const fillRects: Array<[number, number, number, number]> = [];
  const moveTos: Array<[number, number]> = [];
  const arcs: Array<[number, number, number]> = [];
  return {
    save: fn,
    restore: fn,
    beginPath: fn,
    moveTo: (x: number, y: number) => {
      moveTos.push([x, y]);
    },
    lineTo: fn,
    stroke: fn,
    closePath: fn,
    fillRect: (x: number, y: number, width: number, height: number) => {
      fillRects.push([x, y, width, height]);
    },
    strokeRect: fn,
    rect: fn,
    arc: (x: number, y: number, radius: number) => {
      arcs.push([x, y, radius]);
    },
    fill: fn,
    fillText: fn,
    measureText: () => ({ width: 10 }),
    setLineDash: fn,
    get fillStyle() {
      return fillStyles[fillStyles.length - 1] ?? "";
    },
    set fillStyle(value: string | CanvasGradient | CanvasPattern) {
      fillStyles.push(String(value));
    },
    fillStyles,
    fillRects,
    moveTos,
    arcs,
  } as unknown as CanvasRenderingContext2D & {
    fillStyles: string[];
    fillRects: Array<[number, number, number, number]>;
    moveTos: Array<[number, number]>;
    arcs: Array<[number, number, number]>;
  };
}

describe("uPlotAdapter", () => {
  it("draws overlays without a hero graph context", () => {
    const tile: GraphTile = {
      schema_version: "signalforge.graph_tile.v1",
      id: "tile",
      card_id: "tile",
      level: "live",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      renderer: "signalforge.tile.uplot",
      series: [],
      bands: [],
      markers: [],
      events: [],
      diagnostics: {},
      provenance: {},
    };
    const plot = {
      ctx: fakeContext(),
      bbox: { left: 0, top: 0, width: 640, height: 240 },
    } as unknown as uPlot;

    expect(() => drawTileOverlays(plot, tile)).not.toThrow();
  });

  it("falls back to finite tile timestamps when overlay bounds are malformed", () => {
    const tile: GraphTile = {
      schema_version: "signalforge.graph_tile.v1",
      id: "tile",
      card_id: "tile",
      level: "live",
      t0: "not-a-date",
      t1: "also-not-a-date",
      renderer: "signalforge.tile.uplot",
      series: [
        {
          id: "actual",
          label: "Actual",
          role: "actual",
          source: "fixture",
          points: [
            { timestamp: "2024-01-01T00:00:00Z", value: 1 },
            { timestamp: "2024-01-01T00:01:00Z", value: 2 },
          ],
        },
      ],
      bands: [],
      markers: [],
      events: [],
      diagnostics: {},
      provenance: {},
    };
    const ctx = fakeContext();
    const plot = {
      ctx,
      bbox: { left: 0, top: 0, width: 640, height: 240 },
    } as unknown as uPlot;

    expect(() => drawTileOverlays(plot, tile)).not.toThrow();
    expect(ctx.moveTos.length).toBeGreaterThan(0);
  });

  it("uses marker role color for generic marker fills", () => {
    const tile: GraphTile = {
      schema_version: "signalforge.graph_tile.v1",
      id: "tile",
      card_id: "tile",
      level: "live",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      renderer: "signalforge.tile.uplot",
      series: [],
      bands: [],
      markers: [
        {
          id: "evidence",
          timestamp: "2024-01-01T00:00:20Z",
          label: "Evidence",
          role: "evidence",
          kind: "note",
        },
      ],
      events: [],
      diagnostics: {},
      provenance: {},
    };
    const ctx = fakeContext();
    const plot = {
      ctx,
      bbox: { left: 0, top: 0, width: 640, height: 240 },
    } as unknown as uPlot;

    drawTileOverlays(plot, tile);

    expect(ctx.fillStyles).toContain("rgba(176,121,255,0.96)");
  });

  it("keeps padded ranges monotonic when all data fall outside a clamp", () => {
    expect(paddedRange(0.08, [0, 12])({} as uPlot, 20, 24)).toEqual([0, 12]);
    expect(paddedRange(0.08, [0, 12])({} as uPlot, -5, -1)).toEqual([0, 12]);
  });

  it("uses canonical pressure log scales for pressure and pressure-rate series", () => {
    const tile = normalizeGraphTile({
      schema_version: "signalforge.graph_tile.v1",
      id: "tile",
      card_id: "tile",
      level: "live",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      renderer: "signalforge.tile.uplot",
      series: [
        {
          id: "p",
          label: "Pressure",
          role: "actual",
          source: "fixture",
          unit: "mbar",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0.2 }],
        },
        {
          id: "dp",
          label: "Pressure rate",
          role: "actual",
          source: "fixture",
          unit: "mbar/min",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0.05 }],
        },
      ],
      bands: [],
      markers: [],
      events: [],
      diagnostics: {},
      provenance: {},
    });

    const built = uplotData(tile);

    expect(built.series.map((series) => series.scale)).toContain("pressure_log");
    expect(built.series.map((series) => series.scale)).toContain("pressure_rate_log");
    expect(Object.keys(built.scales)).toEqual(expect.arrayContaining(["pressure_log", "pressure_rate_log"]));
  });

  it("keeps explicit linear pressure axes on linear scales", () => {
    const tile = normalizeGraphTile({
      schema_version: "signalforge.graph_tile.v1",
      id: "linear-pressure",
      card_id: "linear-pressure",
      level: "live",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      renderer: "signalforge.tile.uplot",
      series: [
        {
          id: "pressure",
          label: "Pressure",
          role: "actual",
          source: "fixture",
          axis_id: "pressure_mbar",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0 }],
        },
        {
          id: "rate",
          label: "Pressure rate",
          role: "actual",
          source: "fixture",
          axis_id: "pressure_rate",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0 }],
        },
      ],
      bands: [],
      markers: [],
      events: [],
      diagnostics: {},
      provenance: {},
    });

    const built = uplotData(tile);

    expect(built.series.map((series) => series.scale)).toEqual(expect.arrayContaining(["pressure_mbar", "pressure_rate"]));
    expect(Object.keys(built.scales)).toEqual(expect.arrayContaining(["pressure_mbar", "pressure_rate"]));
  });

  it("keeps DUT and AUX role ordering in the uPlot build", () => {
    const tile = normalizeGraphTile({
      schema_version: "signalforge.graph_tile.v1",
      id: "roles",
      card_id: "roles",
      level: "live",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      renderer: "signalforge.tile.uplot",
      series: [
        { id: "target", label: "Target", role: "command", source: "fixture", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 1 }] },
        { id: "actual", label: "Actual", role: "actual", source: "fixture", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 2 }] },
        { id: "load", label: "Load", role: "dut", source: "fixture", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 3 }] },
        { id: "ambient", label: "Ambient", role: "aux", source: "fixture", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 4 }] },
      ],
      bands: [],
      markers: [],
      events: [],
      diagnostics: {},
      provenance: {},
    });

    const built = uplotData(tile);

    expect(tile.series.map((series) => series.role)).toEqual(["command", "actual", "dut", "aux"]);
    expect(built.series.slice(1).map((series) => series.label)).toEqual(["Actual", "Load", "Target", "Ambient"]);
  });

  it("keeps renderer data updates scoped to the currently mounted tile", () => {
    const source = readFileSync(new URL("../src/render/UPlotTileRenderer.tsx", import.meta.url), "utf8");

    expect(source).toContain("const plotTileRef = useRef<GraphTile | null>(null)");
    expect(source).toContain("const plotTile = plotTileRef.current");
    expect(source).toContain("plot.setData(built.data, false)");
    expect(source).toContain("}, [currentTimeMs]);");
    expect(source).not.toContain("}, [currentTimeMs, tile]);");
  });

  it("anchors interlock markers on safety/facility traces instead of command traces", () => {
    const tile = normalizeGraphTile({
      schema_version: "signalforge.graph_tile.v1",
      id: "interlock-anchor",
      card_id: "interlock-anchor",
      level: "live",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      renderer: "signalforge.tile.uplot",
      series: [
        {
          id: "command",
          label: "Command",
          role: "command",
          source: "fixture",
          points: [{ timestamp: "2024-01-01T00:00:30Z", value: 20 }],
        },
        {
          id: "facility_interlock",
          label: "Facility interlock state",
          role: "actual",
          source: "fixture",
          points: [{ timestamp: "2024-01-01T00:00:30Z", value: 80 }],
        },
      ],
      bands: [],
      markers: [
        {
          id: "facility-trip",
          timestamp: "2024-01-01T00:00:30Z",
          label: "Facility interlock",
          kind: "interlock",
          role: "interlock",
        },
      ],
      events: [],
      diagnostics: {},
      provenance: {},
    });
    const ctx = fakeContext();
    const plot = {
      ctx,
      bbox: { left: 0, top: 0, width: 640, height: 100 },
      valToPos: (value: number) => value,
    } as unknown as uPlot;

    drawTileOverlays(plot, tile);

    expect(ctx.arcs).toContainEqual([320, 80, 8]);
    expect(ctx.arcs.some((arc) => arc[0] === 320 && arc[1] === 20)).toBe(false);
  });

  it("clamps the current-time overlay to the plot bounds", () => {
    const tile: GraphTile = {
      schema_version: "signalforge.graph_tile.v1",
      id: "tile",
      card_id: "tile",
      level: "live",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      renderer: "signalforge.tile.uplot",
      series: [],
      bands: [],
      markers: [],
      events: [],
      diagnostics: {},
      provenance: {},
    };
    const beforeCtx = fakeContext();
    const afterCtx = fakeContext();
    const beforePlot = {
      ctx: beforeCtx,
      bbox: { left: 0, top: 0, width: 640, height: 240 },
    } as unknown as uPlot;
    const afterPlot = {
      ctx: afterCtx,
      bbox: { left: 0, top: 0, width: 640, height: 240 },
    } as unknown as uPlot;

    drawTileOverlays(beforePlot, tile, undefined, Date.parse("2023-12-31T23:59:00Z"));
    drawTileOverlays(afterPlot, tile, undefined, Date.parse("2024-01-01T00:02:00Z"));

    expect(beforeCtx.fillRects).toContainEqual([0, 0, 640, 240]);
    expect(beforeCtx.moveTos).toContainEqual([0, 0]);
    expect(afterCtx.fillRects).toContainEqual([640, 0, 0, 240]);
    expect(afterCtx.moveTos).toContainEqual([640, 0]);
  });
});
