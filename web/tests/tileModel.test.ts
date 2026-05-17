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
});
