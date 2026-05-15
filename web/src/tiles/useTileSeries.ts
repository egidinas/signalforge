import { useState, useEffect, useRef } from "react";
import type { GraphTile } from "../types";
import { TileClient, pickTileLevel, type TileLevel } from "./TileClient";
import type { TileAdapter } from "../types";

export type TileState =
  | { status: "loading"; tile: null }
  | { status: "ok"; tile: GraphTile }
  | { status: "error"; tile: null; error: string };

export function useTileSeries(
  adapter: TileAdapter,
  wallId: string,
  cardId: string,
  timeWindowMs: number,
  pollIntervalMs = 5_000
): TileState {
  const [state, setState] = useState<TileState>({ status: "loading", tile: null });
  const clientRef = useRef<TileClient | null>(null);
  if (!clientRef.current) clientRef.current = new TileClient(adapter);

  useEffect(() => {
    const client = clientRef.current!;
    let cancelled = false;
    const level: TileLevel = pickTileLevel(timeWindowMs);

    async function poll() {
      try {
        const tile = await client.fetch(wallId, cardId, level);
        if (!cancelled) setState({ status: "ok", tile });
      } catch (err) {
        if (!cancelled) setState({ status: "error", tile: null, error: String(err) });
      }
    }

    poll();
    if (level === "live") {
      const id = setInterval(poll, pollIntervalMs);
      return () => { cancelled = true; clearInterval(id); };
    }
    return () => { cancelled = true; };
  }, [wallId, cardId, timeWindowMs, pollIntervalMs]);

  return state;
}
