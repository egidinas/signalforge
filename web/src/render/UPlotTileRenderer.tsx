import React, { useEffect, useRef, useCallback } from "react";
import uPlot from "uplot";
import type { GraphTile, HeroGraphModel } from "../types";
import { uplotData, drawTileOverlays } from "./uPlotAdapter";

export type UPlotTileRendererProps = {
  tile: GraphTile;
  heroGraph?: HeroGraphModel;
  height?: number;
  currentTimeMs?: number;
  hoverTimeMs?: number;
  className?: string;
  dataGraphRenderer?: string;
  syncKey?: string;
};

export function UPlotTileRenderer({
  tile,
  heroGraph,
  height = 280,
  currentTimeMs,
  hoverTimeMs,
  className,
  dataGraphRenderer,
  syncKey = "sf-wall",
}: UPlotTileRendererProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const plotRef = useRef<uPlot | null>(null);

  const build = useCallback(() => {
    if (!containerRef.current) return;
    const width = containerRef.current.offsetWidth || 900;
    const built = uplotData(tile, currentTimeMs, width);

    const hooks: uPlot.Hooks.Arrays = {
      draw: [
        (u) => {
          drawTileOverlays(
            u,
            tile,
            heroGraph ?? ({} as HeroGraphModel),
            currentTimeMs,
            hoverTimeMs,
          );
        },
      ],
    };

    const opts: uPlot.Options = {
      width,
      height,
      series: built.series,
      scales: built.scales,
      axes: built.axes,
      hooks,
      cursor: { sync: { key: syncKey } },
    };

    if (plotRef.current) {
      plotRef.current.destroy();
    }
    plotRef.current = new uPlot(opts, built.data, containerRef.current);
  }, [tile, heroGraph, height, currentTimeMs, hoverTimeMs, syncKey]);

  useEffect(() => {
    build();
    const ro = new ResizeObserver(() => {
      if (plotRef.current && containerRef.current) {
        plotRef.current.setSize({ width: containerRef.current.offsetWidth, height });
      }
    });
    if (containerRef.current) ro.observe(containerRef.current);
    return () => {
      ro.disconnect();
      plotRef.current?.destroy();
      plotRef.current = null;
    };
  }, [build, height]);

  return (
    <div
      ref={containerRef}
      className={className}
      data-graph-renderer={dataGraphRenderer ?? tile.renderer ?? "signalforge.tile.uplot"}
      data-graph-tile={tile.tile_id ?? tile.id}
      style={{ width: "100%" }}
    />
  );
}
