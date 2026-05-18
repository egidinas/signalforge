import { DEFAULT_TILE_LEVELS, pickTileLevel, TileClient } from "../src/tiles/TileClient";
import type { TileAdapter } from "../src/types";
import type { GraphTile } from "../src/types";

describe("pickTileLevel", () => {
  it("returns live for ≤5 minutes", () => {
    expect(pickTileLevel(0)).toBe("live");
    expect(pickTileLevel(5 * 60 * 1000)).toBe("live");
  });
  it("returns minute for 5min–6min", () => {
    expect(pickTileLevel(5 * 60 * 1000 + 1)).toBe("minute");
    expect(pickTileLevel(6 * 60 * 1000)).toBe("minute");
  });
  it("returns hour for 6min–60min", () => {
    expect(pickTileLevel(6 * 60 * 1000 + 1)).toBe("hour");
    expect(pickTileLevel(60 * 60 * 1000)).toBe("hour");
  });
  it("returns three_hour for 60min–3h", () => {
    expect(pickTileLevel(60 * 60 * 1000 + 1)).toBe("three_hour");
    expect(pickTileLevel(3 * 60 * 60 * 1000)).toBe("three_hour");
  });
  it("returns day for 3h–24h", () => {
    expect(pickTileLevel(3 * 60 * 60 * 1000 + 1)).toBe("day");
    expect(pickTileLevel(24 * 60 * 60 * 1000)).toBe("day");
  });
  it("returns three_day for >24h", () => {
    expect(pickTileLevel(24 * 60 * 60 * 1000 + 1)).toBe("three_day");
  });
  it("exports the default tile ladder", () => {
    expect(DEFAULT_TILE_LEVELS.map((level) => level.level)).toEqual(["live", "minute", "hour", "three_hour", "day", "three_day"]);
  });
});

function makeTile(cardId: string): GraphTile {
  return {
    schema_version: 1, id: `tile-${cardId}`, card_id: cardId, level: "live",
    t0: "2024-01-01T00:00:00Z", t1: "2024-01-01T00:05:00Z",
    series: [], bands: [], markers: [], events: [],
    diagnostics: { raw_point_count: 0, point_count: 0, decimation: "none", freshness_ms: 0 },
    provenance: {},
  };
}

describe("TileClient", () => {
  it("fetches and caches tiles", async () => {
    let calls = 0;
    const adapter: TileAdapter = {
      fetchTile: async (_w, cardId, _l) => { calls++; return makeTile(cardId); },
    };
    const client = new TileClient(adapter, { ttlMs: 10_000 });
    await client.fetch("wall1", "card-a", "live");
    await client.fetch("wall1", "card-a", "live");
    expect(calls).toBe(1);
  });

  it("deduplicates in-flight requests", async () => {
    let calls = 0;
    const adapter: TileAdapter = {
      fetchTile: async (_w, cardId, _l) => {
        calls++;
        await new Promise((r) => setTimeout(r, 10));
        return makeTile(cardId);
      },
    };
    const client = new TileClient(adapter);
    const [a, b] = await Promise.all([
      client.fetch("wall1", "card-b", "live"),
      client.fetch("wall1", "card-b", "live"),
    ]);
    expect(calls).toBe(1);
    expect(a.card_id).toBe("card-b");
    expect(b.card_id).toBe("card-b");
  });

  it("invalidates cache by wallId", async () => {
    let calls = 0;
    const adapter: TileAdapter = {
      fetchTile: async (_w, cardId, _l) => { calls++; return makeTile(cardId); },
    };
    const client = new TileClient(adapter, { ttlMs: 10_000 });
    await client.fetch("wall1", "card-c", "live");
    client.invalidate("wall1");
    await client.fetch("wall1", "card-c", "live");
    expect(calls).toBe(2);
  });

  it("does not cross-pollute wall cache on invalidate", async () => {
    let calls = 0;
    const adapter: TileAdapter = {
      fetchTile: async (_w, cardId, _l) => { calls++; return makeTile(cardId); },
    };
    const client = new TileClient(adapter, { ttlMs: 10_000 });
    await client.fetch("wall1", "card-d", "live");
    await client.fetch("wall2", "card-d", "live");
    client.invalidate("wall1");
    await client.fetch("wall2", "card-d", "live");
    expect(calls).toBe(2); // wall2 still cached
  });
});
