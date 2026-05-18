import { formatLegendValue, legendReadouts, operatorMarkerLines, rawValueAt, stateAt } from "../src/render/markers";
import { normalizeGraphTile } from "../src/render/tileModel";
import type { GraphMarker, GraphTile, TileSeries } from "../src/types";

describe("markers", () => {
  const tile = (series: TileSeries[], campaign_id = "generic"): GraphTile => ({
      schema_version: "signalforge.graph_tile.v1",
      id: "tile",
      card_id: "tile",
      campaign_id,
      level: "live",
      t0: "1970-01-01T00:00:00.000Z",
      t1: "2024-01-01T04:00:00.000Z",
      renderer: "signalforge.tile.uplot",
      series,
      bands: [],
      markers: [],
      events: [],
      diagnostics: {},
      provenance: {},
    });

  it("keeps epoch-zero legend readouts", () => {
    const graphTile = tile([{
      id: "temp",
      label: "Temperature",
      role: "actual",
      source: "fixture",
      axis_id: "temperature_c",
      unit: "degC",
      points: [{ timestamp: "1970-01-01T00:00:00.000Z", value: 21.25 }],
    }]);

    const readouts = legendReadouts(graphTile, [{ id: "temp", label: "Temperature" }], 0);

    expect(readouts.get("temp")).toBe("21.3 degC");
  });

  it("treats placeholder units as missing in legend readouts", () => {
    expect(formatLegendValue({
      id: "temp",
      label: "Temperature",
      role: "actual",
      source: "fixture",
      axis_id: "temperature_c",
      unit: "_",
      points: [],
    }, 21.25)).toBe("21.3 degC");
  });

  it("formats normalized loose Celsius units through the canonical temperature axis", () => {
    const graphTile = normalizeGraphTile(tile([{
      id: "temp",
      label: "Temperature",
      role: "actual",
      source: "fixture",
      unit: "deg C",
      points: [{ timestamp: "1970-01-01T00:00:00.000Z", value: 21.25 }],
    }]));

    const readouts = legendReadouts(graphTile, [{ id: "temp", label: "Temperature" }], 0);

    expect(graphTile.series[0].axis_id).toBe("temperature_c");
    expect(readouts.get("temp")).toBe("21.3 degC");
  });

  it("does not report interpolated values inside command center visual gaps", () => {
    const graphTile = tile([{
      id: "actual",
      label: "Actual",
      role: "actual",
      source: "fixture",
      axis_id: "temperature_c",
      unit: "degC",
      points: [
        { timestamp: "2024-01-01T00:00:00.000Z", value: 20 },
        { timestamp: "2024-01-01T04:00:00.000Z", value: 40 },
      ],
    }], "command_center_fat");

    const readouts = legendReadouts(graphTile, [{ id: "actual", label: "Actual" }], Date.parse("2024-01-01T02:00:00.000Z"));

    expect(readouts.has("actual")).toBe(false);
  });

  it("keeps fresh command-center readouts just after the latest sample", () => {
    const series: TileSeries = {
      id: "actual",
      label: "Actual",
      role: "actual",
      source: "fixture",
      axis_id: "temperature_c",
      unit: "degC",
      points: [
        { timestamp: "2024-01-01T00:00:00.000Z", value: 20 },
        { timestamp: "2024-01-01T00:10:00.000Z", value: 21 },
      ],
    };
    const graphTile = tile([series], "command_center_fat");
    const freshCursor = Date.parse("2024-01-01T00:11:00.000Z");
    const staleCursor = Date.parse("2024-01-01T02:11:00.000Z");

    expect(rawValueAt(series, freshCursor, graphTile)).toBe(21);
    expect(legendReadouts(graphTile, [{ id: "actual", label: "Actual" }], freshCursor).get("actual")).toBe("21.0 degC");
    expect(rawValueAt(series, staleCursor, graphTile)).toBeUndefined();
  });

  it("suppresses pressure legend values the plot treats as invalid", () => {
    const graphTile = tile([{
      id: "pressure",
      label: "Pressure",
      role: "actual",
      source: "fixture",
      axis_id: "pressure_log",
      points: [{ timestamp: "2024-01-01T00:00:00.000Z", value: 0 }],
    }]);

    const readouts = legendReadouts(graphTile, [{ id: "pressure", label: "Pressure" }], Date.parse("2024-01-01T00:00:00.000Z"));

    expect(readouts.has("pressure")).toBe(false);
  });

  it("keeps zero readouts on linear pressure axes", () => {
    const graphTile = tile([
      {
        id: "pressure",
        label: "Pressure",
        role: "actual",
        source: "fixture",
        axis_id: "pressure_mbar",
        points: [{ timestamp: "2024-01-01T00:00:00.000Z", value: 0 }],
      },
      {
        id: "pressure_rate",
        label: "Pressure rate",
        role: "actual",
        source: "fixture",
        axis_id: "pressure_rate",
        points: [{ timestamp: "2024-01-01T00:00:00.000Z", value: 0 }],
      },
    ]);

    const readouts = legendReadouts(graphTile, [
      { id: "pressure", label: "Pressure" },
      { id: "pressure_rate", label: "Pressure rate" },
    ], Date.parse("2024-01-01T00:00:00.000Z"));

    expect(readouts.get("pressure")).toBe("0 mbar");
    expect(readouts.get("pressure_rate")).toBe("0 mbar/min");
  });

  it("matches pressure legend readouts to log-space plotted interpolation", () => {
    const series: TileSeries = {
      id: "pressure",
      label: "Pressure",
      role: "actual",
      source: "fixture",
      axis_id: "pressure_log",
      points: [
        { timestamp: "2024-01-01T00:00:00.000Z", value: 1 },
        { timestamp: "2024-01-01T00:20:00.000Z", value: 100 },
      ],
    };
    const graphTile = tile([series]);
    const midpoint = Date.parse("2024-01-01T00:10:00.000Z");

    expect(rawValueAt(series, midpoint, graphTile)).toBeCloseTo(10, 5);
    expect(legendReadouts(graphTile, [{ id: "pressure", label: "Pressure" }], midpoint).get("pressure")).toBe("10.0 mbar");
  });

  it("does not bridge through invalid pressure samples", () => {
    const series: TileSeries = {
      id: "pressure",
      label: "Pressure",
      role: "actual",
      source: "fixture",
      axis_id: "pressure_log",
      points: [
        { timestamp: "2024-01-01T00:00:00.000Z", value: 1 },
        { timestamp: "2024-01-01T00:10:00.000Z", value: 0 },
        { timestamp: "2024-01-01T00:20:00.000Z", value: 100 },
      ],
    };

    expect(rawValueAt(series, Date.parse("2024-01-01T00:15:00.000Z"), tile([series]))).toBeUndefined();
  });

  it("keeps finite pressure readouts immediately before invalid sentinels", () => {
    const series: TileSeries = {
      id: "pressure",
      label: "Pressure",
      role: "actual",
      source: "fixture",
      axis_id: "pressure_log",
      points: [
        { timestamp: "2024-01-01T00:00:00.000Z", value: 1 },
        { timestamp: "2024-01-01T00:05:00.000Z", value: 10 },
        { timestamp: "2024-01-01T00:10:00.000Z", value: 0 },
        { timestamp: "2024-01-01T00:20:00.000Z", value: 100 },
      ],
    };
    const graphTile = tile([series]);
    const finiteBoundary = Date.parse("2024-01-01T00:05:00.000Z");
    const gapSentinel = Date.parse("2024-01-01T00:10:00.000Z");

    expect(rawValueAt(series, finiteBoundary, graphTile)).toBe(10);
    expect(legendReadouts(graphTile, [{ id: "pressure", label: "Pressure" }], finiteBoundary).get("pressure")).toBe("10.0 mbar");
    expect(rawValueAt(series, gapSentinel, graphTile)).toBeUndefined();
  });

  it("resolves state span readouts through value tables", () => {
    const series: TileSeries = {
      id: "phase",
      label: "Phase",
      role: "state",
      source: "fixture",
      render_kind: "swimlane",
      value_table: { "2": "RESET_READY" },
      spans: [{ start: "2024-01-01T00:00:00.000Z", end: "2024-01-01T00:05:00.000Z", value: 2 }],
    };

    expect(stateAt(series, Date.parse("2024-01-01T00:02:00.000Z"))).toBe("RESET_READY");
    expect(legendReadouts(tile([series]), [{ id: "phase", label: "Phase" }], Date.parse("2024-01-01T00:02:00.000Z")).get("phase")).toBe("RESET_READY");
  });

  it("keeps unknown operator-interaction markers out of the READY bucket", () => {
    const marker = {
      id: "op-abort",
      timestamp: "2024-01-01T00:00:00.000Z",
      label: "Operator abort",
      role: "operator_interaction",
      kind: "operator_abort",
    } as GraphMarker;

    expect(operatorMarkerLines(marker)[0]).toBe("ABORT");
    expect(operatorMarkerLines({ ...marker, kind: "operator_reset_ready" }, true)[0]).toBe("RDY");
  });
});
