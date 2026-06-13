import { useEffect, useMemo, useRef, useState } from "react";
import { hasRenderableGraphTile, retainGraphTile, type RetainGraphTileOptions } from "../render/tileModel";

export type ViewportRetainedGraphTileOptions = RetainGraphTileOptions & {
  readonly rootMargin?: string;
  readonly threshold?: number | number[];
};

export function useViewportRetainedGraphTile<TileLike>(
  tile: TileLike,
  opts: ViewportRetainedGraphTileOptions = {},
) {
  const rootRef = useRef<HTMLDivElement | null>(null);
  const lastServedRef = useRef<TileLike | null>(null);
  const rootMargin = opts.rootMargin || "160px";
  const threshold = opts.threshold ?? 0;
  const [inViewport, setInViewport] = useState(true);

  useEffect(() => {
    const el = rootRef.current;
    if (!el || typeof IntersectionObserver === "undefined") {
      setInViewport(true);
      return undefined;
    }

    const observer = new IntersectionObserver((entries) => {
      const visible = entries.some((entry) => entry.isIntersecting);
      setInViewport(visible);
      if (!visible) lastServedRef.current = null;
    }, { root: null, rootMargin, threshold });
    observer.observe(el);
    return () => observer.disconnect();
  }, [rootMargin, threshold]);

  const retainedTile = useMemo(
    () => retainGraphTile(tile, lastServedRef.current, { inViewport }),
    [tile, inViewport],
  );

  useEffect(() => {
    if (hasRenderableGraphTile(tile)) {
      lastServedRef.current = tile;
    } else if (!inViewport) {
      lastServedRef.current = null;
    }
  }, [tile, inViewport]);

  return { ref: rootRef, tile: retainedTile, retained: retainedTile !== tile, inViewport };
}
