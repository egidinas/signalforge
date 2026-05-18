import type { GraphTile, TileAdapter } from "../types";

export type TileLevel = "live" | "minute" | "hour" | "three_hour" | "day" | "three_day";

export type TileLevelSpec = {
  level: TileLevel;
  label: string;
  timeWindowMs: number;
};

export const DEFAULT_TILE_LEVELS: TileLevelSpec[] = [
  { level: "live", label: "live 90s", timeWindowMs: 90_000 },
  { level: "minute", label: "history 6m", timeWindowMs: 6 * 60_000 },
  { level: "hour", label: "history 60m", timeWindowMs: 60 * 60_000 },
  { level: "three_hour", label: "history 3h", timeWindowMs: 3 * 60 * 60_000 },
  { level: "day", label: "history 24h", timeWindowMs: 24 * 60 * 60_000 },
  { level: "three_day", label: "history 3d", timeWindowMs: 3 * 24 * 60 * 60_000 },
];

export function pickTileLevel(timeWindowMs: number): TileLevel {
  if (timeWindowMs <= 5 * 60_000) return "live";
  if (timeWindowMs <= 6 * 60_000) return "minute";
  if (timeWindowMs <= 60 * 60_000) return "hour";
  if (timeWindowMs <= 3 * 60 * 60_000) return "three_hour";
  if (timeWindowMs <= 24 * 60 * 60_000) return "day";
  return "three_day";
}

type CacheEntry = { tile: GraphTile; fetchedAt: number };
type InFlight = Promise<GraphTile>;

export class TileClient {
  private cache = new Map<string, CacheEntry>();
  private inflight = new Map<string, InFlight>();
  private ttlMs: number;

  constructor(private adapter: TileAdapter, opts: { ttlMs?: number } = {}) {
    this.ttlMs = opts.ttlMs ?? 30_000;
  }

  private cacheKey(wallId: string, cardId: string, level: TileLevel): string {
    return `${wallId}/${cardId}@${level}`;
  }

  async fetch(wallId: string, cardId: string, level: TileLevel): Promise<GraphTile> {
    const key = this.cacheKey(wallId, cardId, level);
    const cached = this.cache.get(key);
    if (cached && Date.now() - cached.fetchedAt < this.ttlMs) return cached.tile;
    const existing = this.inflight.get(key);
    if (existing) return existing;
    const promise = this.adapter.fetchTile(wallId, cardId, level).then((tile) => {
      this.cache.set(key, { tile, fetchedAt: Date.now() });
      this.inflight.delete(key);
      return tile;
    }).catch((err) => {
      this.inflight.delete(key);
      throw err;
    });
    this.inflight.set(key, promise);
    return promise;
  }

  fetchForViewport(wallId: string, cardId: string, timeWindowMs: number): Promise<GraphTile> {
    return this.fetch(wallId, cardId, pickTileLevel(timeWindowMs));
  }

  invalidate(wallId?: string): void {
    if (!wallId) { this.cache.clear(); return; }
    for (const key of this.cache.keys()) {
      if (key.startsWith(`${wallId}/`)) this.cache.delete(key);
    }
  }
}
