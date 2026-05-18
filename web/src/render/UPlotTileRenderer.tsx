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
  const plotTileRef = useRef<GraphTile | null>(null);
  const heroGraphRef = useRef<HeroGraphModel | undefined>(heroGraph);
  const currentTimeRef = useRef<number | undefined>(currentTimeMs);
  const hoverTimeRef = useRef<number | undefined>(hoverTimeMs);

  useEffect(() => {
    heroGraphRef.current = heroGraph;
    plotRef.current?.redraw();
  }, [heroGraph]);

  useEffect(() => {
    currentTimeRef.current = currentTimeMs;
    const plot = plotRef.current;
    const plotTile = plotTileRef.current;
    if (!plot || !plotTile) return;
    const width = plot.width || containerRef.current?.offsetWidth || 900;
    const built = uplotData(plotTile, currentTimeMs, width);
    plot.setData(built.data, false);
    plot.redraw();
  }, [currentTimeMs]);

  useEffect(() => {
    hoverTimeRef.current = hoverTimeMs;
    plotRef.current?.redraw();
  }, [hoverTimeMs]);

  const build = useCallback(() => {
    if (!containerRef.current) return;
    const width = containerRef.current.offsetWidth || 900;
    const built = uplotData(tile, currentTimeRef.current, width);

    const hooks: uPlot.Hooks.Arrays = {
      draw: [
        (u) => {
          drawTileOverlays(
            u,
            tile,
            heroGraphRef.current,
            currentTimeRef.current,
            hoverTimeRef.current,
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
    plotTileRef.current = tile;
  }, [tile, height, syncKey]);

  useEffect(() => {
    build();
    const ro = new ResizeObserver(() => {
      const plot = plotRef.current;
      const plotTile = plotTileRef.current;
      const container = containerRef.current;
      if (!plot || !plotTile || !container) return;
      const width = container.offsetWidth || plot.width || 900;
      plot.setSize({ width, height });
      const built = uplotData(plotTile, currentTimeRef.current, width);
      plot.setData(built.data, false);
      plot.redraw();
    });
    if (containerRef.current) ro.observe(containerRef.current);
    return () => {
      ro.disconnect();
      plotRef.current?.destroy();
      plotRef.current = null;
      plotTileRef.current = null;
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
