import {
  CANONICAL_TILE_RENDERER,
  normalizeGraphTile,
  renderSeriesFromGraphTile,
  seriesRoleMeta,
} from "../src/render/tileModel";

describe("tileModel", () => {
  it("normalizes loose graph tiles into the canonical uPlot renderer contract", () => {
    const tile = normalizeGraphTile({
      id: "controller-temp",
      card_id: "controller-temp",
      series: [
        {
          series_id: "target",
          role: "cmd",
          unit: "deg C",
          source: { device_id: "tec-a", param_id: 3000, instance: 1 },
          history: { ts: [1_700_000_000_000, null, ""], v: [18.5, null, ""] },
        },
        {
          id: "empty",
          role: "actual",
          points: [],
        },
      ],
    });

    expect(tile.renderer).toBe(CANONICAL_TILE_RENDERER);
    expect(tile.series).toHaveLength(1);
    expect(tile.series[0].role).toBe("command");
    expect(tile.series[0].axis_id).toBe("temperature_c");
    expect(tile.series[0].points?.[0]).toEqual({
      timestamp: "2023-11-14T22:13:20.000Z",
      value: 18.5,
    });
    expect(tile.diagnostics).toMatchObject({
      status: "ok",
      point_count: 1,
      renderer: CANONICAL_TILE_RENDERER,
      series_count: 1,
    });
  });

  it("orders rendered series by shared role metadata", () => {
    const rendered = renderSeriesFromGraphTile({
      id: "tile-a",
      card_id: "tile-a",
      t0: "2024-01-01T00:00:00Z",
      t1: "2024-01-01T00:01:00Z",
      series: [
        { id: "actual", label: "Actual", role: "actual", source: "actual", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 1 }] },
        { id: "target", label: "Target", role: "cmd", source: "target", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 2 }] },
      ],
    });

    expect(seriesRoleMeta("cmd").rank).toBeLessThan(seriesRoleMeta("actual").rank);
    expect(rendered.map((series) => series.key)).toEqual(["target", "actual"]);
    expect(rendered[0].history).toEqual({ ts: [1_704_067_200_000], v: [2], q: ["ok"] });
  });

  it("preserves structured source_ref metadata for rendered series", () => {
    const tile = normalizeGraphTile({
      id: "source-ref",
      card_id: "source-ref",
      series: [
        {
          id: "pressure",
          label: "Pressure",
          role: "actual",
          source: "fixture",
          source_ref: {
            device_id: "tec-1",
            param_id: "1042",
            instance: "ch1",
            signal_id: "pressure.actual",
          },
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0.02 }],
        },
      ],
    });

    expect(tile.series[0] as typeof tile.series[0] & { source_obj?: unknown }).toMatchObject({
      source: "device=tec-1 param=1042 instance=ch1 signal=pressure.actual endpoint=",
      source_obj: {
        device_id: "tec-1",
        param_id: "1042",
        instance: "ch1",
        signal_id: "pressure.actual",
      },
    });
    expect(renderSeriesFromGraphTile(tile)[0]).toMatchObject({
      deviceId: "tec-1",
      paramId: "1042",
      instance: "ch1",
      signalId: "pressure.actual",
    });
  });

  it("derives fallback time bounds from span-only series", () => {
    const tile = normalizeGraphTile({
      id: "state",
      card_id: "state",
      series: [
        {
          id: "state",
          label: "State",
          role: "actual",
          source: "state",
          spans: [
            {
              start: "2024-01-01T00:00:10Z",
              end: "2024-01-01T00:00:20Z",
              state: "on",
            },
          ],
        },
      ],
    });

    expect(tile.series).toHaveLength(1);
    expect(tile.t0).toBe("2024-01-01T00:00:10.000Z");
    expect(tile.t1).toBe("2024-01-01T00:00:20.000Z");
  });

  it("derives fallback time bounds from overlay-only tiles", () => {
    const tile = normalizeGraphTile({
      id: "overlays",
      card_id: "overlays",
      bands: [{ id: "dwell", start: "2024-01-01T00:00:10Z", end: "2024-01-01T00:00:20Z" }],
      markers: [{ id: "mark", label: "Mark", kind: "gate", timestamp: "2024-01-01T00:00:05Z" }],
      events: [{ id: "event", label: "Event", kind: "note", timestamp: "2024-01-01T00:00:25Z" }],
      series: [],
    });

    expect(tile.series).toHaveLength(0);
    expect(tile.t0).toBe("2024-01-01T00:00:05.000Z");
    expect(tile.t1).toBe("2024-01-01T00:00:25.000Z");
  });

  it("preserves source diagnostics status when data are present", () => {
    const tile = normalizeGraphTile({
      id: "stale-tile",
      card_id: "stale-tile",
      diagnostics: { status: "stale", raw_point_count: 12 },
      series: [
        {
          id: "temperature",
          label: "Temperature",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 21.5 }],
        },
      ],
    });

    expect(tile.diagnostics.status).toBe("stale");
    expect(tile.diagnostics.point_count).toBe(1);
    expect(tile.diagnostics.series_count).toBe(1);
    expect(tile.diagnostics.raw_point_count).toBe(12);
  });

  it("preserves DUT and AUX render roles through normalization", () => {
    const tile = normalizeGraphTile({
      id: "role-contract",
      card_id: "role-contract",
      series: [
        {
          id: "load",
          label: "Load",
          role: "dut",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 12 }],
        },
        {
          id: "ambient",
          label: "Ambient",
          role: "aux",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 19 }],
        },
      ],
    });

    expect(tile.series.map((series) => series.role)).toEqual(["dut", "aux"]);
    const rendered = renderSeriesFromGraphTile(tile);
    expect(rendered.map((series) => series.seriesRole)).toEqual(["load", "ambient"].map((_, index) => tile.series[index].role));
    expect(rendered.map((series) => series.roleRank)).toEqual([
      seriesRoleMeta("dut").rank,
      seriesRoleMeta("aux").rank,
    ]);
  });

  it("infers pressure axis ids from loose units and labels", () => {
    const tile = normalizeGraphTile({
      id: "pressure",
      card_id: "pressure",
      series: [
        { id: "chamber", label: "Chamber pressure", unit: "mbar", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0.02 }] },
        { id: "pumpdown", label: "Pumpdown rate", unit: "mbar/min", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0.1 }] },
        { id: "supply", label: "Supply", unit: "bar", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 3.4 }] },
        { id: "vacuum", label: "Vacuum level", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0.01 }] },
      ],
    });

    expect(tile.series.map((series) => series.axis_id)).toEqual([
      "pressure_log",
      "pressure_rate_log",
      "pressure_bar",
      "pressure_log",
    ]);
  });

  it("infers temperature axes from degree-symbol Celsius units", () => {
    const tile = normalizeGraphTile({
      id: "degree-symbol",
      card_id: "degree-symbol",
      series: [
        { id: "temperature", label: "Temperature", unit: "°C", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 21 }] },
      ],
    });

    expect(tile.series[0].axis_id).toBe("temperature_c");
  });

  it("infers semantic axes from canonical series id fields", () => {
    const tile = normalizeGraphTile({
      id: "semantic",
      card_id: "semantic",
      series: [
        { series_id: "trace.tvac_pressure", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0.02 }] },
        { key: "pump.cycle_counter", points: [{ timestamp: "2024-01-01T00:00:00Z", value: 12 }] },
      ],
    });

    expect(tile.series.map((series) => series.axis_id)).toEqual(["pressure_log", "counter"]);
  });

  it("keeps counter role semantics ahead of pressure-like names", () => {
    const tile = normalizeGraphTile({
      id: "counter",
      card_id: "counter",
      series: [
        {
          id: "pressure_faults",
          label: "Pressure fault counter",
          role: "counter",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 0 }],
        },
      ],
    });

    expect(tile.series[0].axis_id).toBe("counter");
    expect(renderSeriesFromGraphTile(tile)[0].history.v).toEqual([0]);
  });

  it("keeps legacy seriesRole counter semantics for opaque ids", () => {
    const tile = normalizeGraphTile({
      id: "counter-role",
      card_id: "counter-role",
      series: [
        {
          id: "packets",
          seriesRole: "counter",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 5 }],
        },
      ],
    });

    expect(tile.series[0].role).toBe("counter");
    expect(tile.series[0].axis_id).toBe("counter");
  });

  it("keeps explicit counter roles ahead of pressure units", () => {
    const tile = normalizeGraphTile({
      id: "counter-unit",
      card_id: "counter-unit",
      series: [
        {
          id: "pump_cycles",
          role: "counter",
          unit: "mbar",
          points: [{ timestamp: "2024-01-01T00:00:00Z", value: 7 }],
        },
      ],
    });

    expect(tile.series[0].axis_id).toBe("counter");
  });
});
