import React, { useEffect, useRef, useCallback } from "react";
import uPlot from "uplot";
import type { GraphTile, HeroGraphModel } from "../types";
import { uplotData, drawTileOverlays } from "./uPlotAdapter";
import { measuredElementSize } from "./tileModel";

export type UPlotTileRendererProps = {
  tile: GraphTile;
  heroGraph?: HeroGraphModel;
  height?: number;
  currentTimeMs?: number;
  hoverTimeMs?: number;
  className?: string;
  dataGraphRenderer?: string;
  syncKey?: string;
  autoY?: boolean;
  yRange?: [number, number] | null;
  viewToken?: number;
  fillContainer?: boolean;
  minWidth?: number;
  minHeight?: number;
};

function rendererSize(
  el: HTMLElement | null,
  fallbackHeight: number,
  fillContainer: boolean,
  minWidth: number,
  minHeight: number,
) {
  const measured = measuredElementSize(el);
  return {
    width: Math.max(minWidth, measured.width || minWidth),
    height: fillContainer ? Math.max(minHeight, measured.height || fallbackHeight) : fallbackHeight,
  };
}

export function UPlotTileRenderer({
  tile,
  heroGraph,
  height = 280,
  currentTimeMs,
  hoverTimeMs,
  className,
  dataGraphRenderer,
  syncKey = "sf-wall",
  autoY = true,
  yRange = null,
  viewToken = 0,
  fillContainer = false,
  minWidth = 180,
  minHeight = 160,
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
    const { width } = rendererSize(containerRef.current, height, fillContainer, minWidth, minHeight);
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
    const { width, height: plotHeight } = rendererSize(containerRef.current, height, fillContainer, minWidth, minHeight);
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
      height: plotHeight,
      series: built.series,
      scales: {
        ...built.scales,
        ...(autoY && !yRange ? {} : (() => {
          const yAxis = built.axes.find((axis, idx) => idx > 0 && axis?.show !== false && axis?.scale);
          const scaleName = yAxis?.scale ?? "generic_numeric";
          return {
            [scaleName]: {
              ...(built.scales[scaleName] || {}),
              ...(yRange ? { range: () => yRange } : {}),
            },
          };
        })()),
      },
      axes: built.axes,
      hooks,
      cursor: { sync: { key: syncKey } },
    };

    if (plotRef.current) {
      plotRef.current.destroy();
    }
    plotRef.current = new uPlot(opts, built.data, containerRef.current);
    plotTileRef.current = tile;
  }, [tile, height, syncKey, autoY, yRange?.[0], yRange?.[1], viewToken, fillContainer, minWidth, minHeight]);

  useEffect(() => {
    build();
    const ro = new ResizeObserver(() => {
      const plot = plotRef.current;
      const plotTile = plotTileRef.current;
      const container = containerRef.current;
      if (!plot || !plotTile || !container) return;
      const { width, height: plotHeight } = rendererSize(container, height, fillContainer, minWidth, minHeight);
      plot.setSize({ width, height: plotHeight });
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
  }, [build, height, fillContainer, minWidth, minHeight]);

  return (
    <div
      ref={containerRef}
      className={className}
      data-graph-renderer={dataGraphRenderer ?? tile.renderer ?? "signalforge.tile.uplot"}
      data-graph-fill={fillContainer ? "container" : "fixed"}
      data-graph-tile={tile.tile_id ?? tile.id}
      style={{
        width: "100%",
        minWidth: 0,
        ...(fillContainer ? { height: "100%", minHeight } : undefined),
      }}
    />
  );
}
