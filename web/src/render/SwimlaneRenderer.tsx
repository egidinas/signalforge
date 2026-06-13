import React, { useEffect, useRef, useCallback } from "react";
import type { GraphTile } from "../types";
import { stateBlocks } from "./uPlotAdapter";

// Default dark-theme palette (value → fill color).
// Matches the shared consumer design tokens so the default looks right out of the box.
const DEFAULT_STATE_PALETTE: Record<number, string> = {
  0: "#1b2a3a",
};
const DEFAULT_STATE_POSITIVE = ["#00b89c", "#4488ff", "#ffb700", "#24d5ff", "#c878ff"];
const DEFAULT_STATE_NEGATIVE = "#d63031";
const DEFAULT_RAIL_BG = "#060d14";
const DEFAULT_BORDER = "#1c2c3c";
const DEFAULT_TEXT_DIM = "#6b8ba4";
const DEFAULT_TEXT_BRIGHT = "#dce9f4";
const DEFAULT_LABEL_RAIL = 132;
const LANE_H = 36;
const BAR_H = 20;
const BAR_Y_OFFSET = (LANE_H - BAR_H) / 2;
const LABEL_FONT = '11px ui-monospace, "SF Mono", monospace';
const STATE_FONT = '10px ui-monospace, "SF Mono", monospace';
const MIN_BLOCK_PX_FOR_LABEL = 28;
const FLAP_REPEAT_GUARD_PX = 56;
const TIME_GRID_STEPS = 10;

export type SwimlaneColorForValue = (value: number | null | undefined) => string;
export type SwimlaneTextForFill = (fill: string) => string;

function defaultColorForValue(value: number | null | undefined): string {
  if (value === null || value === undefined || !Number.isFinite(value)) return DEFAULT_STATE_PALETTE[0];
  if (value === 0) return DEFAULT_STATE_PALETTE[0];
  if (value < 0) return DEFAULT_STATE_NEGATIVE;
  const idx = Math.abs(Math.trunc(value) - 1) % DEFAULT_STATE_POSITIVE.length;
  return DEFAULT_STATE_POSITIVE[idx];
}

function defaultTextForFill(fill: string): string {
  const dark = fill === DEFAULT_STATE_PALETTE[0] || fill.toLowerCase() === DEFAULT_STATE_NEGATIVE.toLowerCase();
  return dark ? DEFAULT_TEXT_BRIGHT : "#06120a";
}

// Truncate text to fit in maxPx given approx charWidth.
function truncateLabel(label: string, maxPx: number, charWidthPx = 6.2): string {
  const maxChars = Math.floor(maxPx / charWidthPx);
  if (label.length <= maxChars) return label;
  if (maxChars <= 1) return "";
  if (maxChars <= 3) return label.slice(0, maxChars);
  return label.slice(0, maxChars - 1) + "…";
}

function drawSwimlanes(
  canvas: HTMLCanvasElement,
  tile: GraphTile,
  width: number,
  height: number,
  colorForValue: SwimlaneColorForValue,
  textForFill: SwimlaneTextForFill,
  currentTimeMs?: number,
  hoverTimeMs?: number,
  labelRailPx = DEFAULT_LABEL_RAIL,
) {
  const dpr = window.devicePixelRatio || 1;
  canvas.width = Math.round(width * dpr);
  canvas.height = Math.round(height * dpr);
  canvas.style.width = `${width}px`;
  canvas.style.height = `${height}px`;

  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.scale(dpr, dpr);

  const start = Date.parse(tile.t0);
  const end = Date.parse(tile.t1);
  const span = Math.max(1, end - start);
  const plotLeft = labelRailPx;
  const plotWidth = Math.max(1, width - labelRailPx);

  // Background
  ctx.fillStyle = DEFAULT_RAIL_BG;
  ctx.fillRect(0, 0, width, height);

  // Label rail border
  ctx.strokeStyle = DEFAULT_BORDER;
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(plotLeft + 0.5, 0);
  ctx.lineTo(plotLeft + 0.5, height);
  ctx.stroke();

  // Plot area background
  ctx.fillStyle = "#0a1622";
  ctx.fillRect(plotLeft, 0, plotWidth, height);

  // Faint time grid
  ctx.strokeStyle = "rgba(83,112,140,0.18)";
  ctx.lineWidth = 1;
  ctx.setLineDash([]);
  for (let i = 0; i <= TIME_GRID_STEPS; i++) {
    const x = plotLeft + (i / TIME_GRID_STEPS) * plotWidth;
    ctx.beginPath();
    ctx.moveTo(Math.round(x) + 0.5, 0);
    ctx.lineTo(Math.round(x) + 0.5, height);
    ctx.stroke();
  }

  // Series — render each lane
  const stateSeries = tile.series.filter((s) => s.render_kind === "swimlane" || s.kind === "state" || (s.spans && s.spans.length > 0));

  stateSeries.forEach((series, laneIdx) => {
    const laneY = laneIdx * LANE_H;
    const barY = laneY + BAR_Y_OFFSET;

    // Lane separator
    if (laneIdx > 0) {
      ctx.strokeStyle = DEFAULT_BORDER;
      ctx.lineWidth = 1;
      ctx.setLineDash([]);
      ctx.beginPath();
      ctx.moveTo(0, laneY + 0.5);
      ctx.lineTo(width, laneY + 0.5);
      ctx.stroke();
    }

    // Signal name in rail
    ctx.font = LABEL_FONT;
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.fillStyle = DEFAULT_TEXT_DIM;
    const labelMaxPx = labelRailPx - 10;
    const labelText = truncateLabel(series.label, labelMaxPx);
    ctx.fillText(labelText, 6, laneY + LANE_H / 2);

    // Compute blocks
    const blocks = stateBlocks(series, start, span);

    // Track last label placement per state value to avoid repetition
    const lastLabelX = new Map<number, number>();

    // Draw colored bars
    blocks.forEach((block) => {
      const bx = plotLeft + (block.left / 100) * plotWidth;
      const bw = Math.max(1, (block.width / 100) * plotWidth);
      const fill = colorForValue(block.value);
      ctx.fillStyle = fill;
      ctx.fillRect(bx, barY, bw, BAR_H);
    });

    // Draw state labels on top of bars
    blocks.forEach((block) => {
      const bx = plotLeft + (block.left / 100) * plotWidth;
      const bw = Math.max(1, (block.width / 100) * plotWidth);
      if (bw < MIN_BLOCK_PX_FOR_LABEL) return;

      const fill = colorForValue(block.value);
      const lastX = lastLabelX.get(block.value);
      if (lastX !== undefined && bx - lastX < FLAP_REPEAT_GUARD_PX) return;

      const availableWidth = bw - 8;
      const labelText = truncateLabel(block.label, availableWidth);
      if (!labelText) return;

      ctx.font = STATE_FONT;
      ctx.textAlign = "left";
      ctx.textBaseline = "middle";
      ctx.fillStyle = textForFill(fill);
      ctx.fillText(labelText, bx + 4, laneY + LANE_H / 2);

      lastLabelX.set(block.value, bx);
    });
  });

  // Now-cursor (shaded future area)
  if (currentTimeMs !== undefined && Number.isFinite(currentTimeMs)) {
    const x = plotLeft + ((currentTimeMs - start) / span) * plotWidth;
    const cx = Math.max(plotLeft, Math.min(plotLeft + plotWidth, x));
    if (cx < plotLeft + plotWidth) {
      ctx.fillStyle = "rgba(3,7,12,0.42)";
      ctx.fillRect(cx, 0, plotLeft + plotWidth - cx, height);
    }
    ctx.strokeStyle = "rgba(242,247,255,0.85)";
    ctx.setLineDash([3, 3]);
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(cx + 0.5, 0);
    ctx.lineTo(cx + 0.5, height);
    ctx.stroke();
    ctx.setLineDash([]);
  }

  // Hover cursor
  if (hoverTimeMs !== undefined && Number.isFinite(hoverTimeMs)) {
    const x = plotLeft + ((hoverTimeMs - start) / span) * plotWidth;
    if (x >= plotLeft && x <= plotLeft + plotWidth) {
      ctx.strokeStyle = "rgba(255,216,95,0.95)";
      ctx.setLineDash([]);
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(x + 0.5, 0);
      ctx.lineTo(x + 0.5, height);
      ctx.stroke();
    }
  }
}

export type SwimlaneRendererProps = {
  tile: GraphTile;
  height?: number;
  currentTimeMs?: number;
  hoverTimeMs?: number;
  className?: string;
  colorForValue?: SwimlaneColorForValue;
  textForFill?: SwimlaneTextForFill;
  labelRailPx?: number;
};

export function SwimlaneRenderer({
  tile,
  height,
  currentTimeMs,
  hoverTimeMs,
  className,
  colorForValue = defaultColorForValue,
  textForFill = defaultTextForFill,
  labelRailPx = DEFAULT_LABEL_RAIL,
}: SwimlaneRendererProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const stateSeries = tile.series.filter(
    (s) => s.render_kind === "swimlane" || s.kind === "state" || (s.spans && s.spans.length > 0),
  );
  const computedHeight = height ?? Math.max(LANE_H, stateSeries.length * LANE_H);

  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;
    const w = container.offsetWidth || 400;
    drawSwimlanes(canvas, tile, w, computedHeight, colorForValue, textForFill, currentTimeMs, hoverTimeMs, labelRailPx);
  }, [tile, computedHeight, colorForValue, textForFill, currentTimeMs, hoverTimeMs, labelRailPx]);

  useEffect(() => {
    draw();
  }, [draw]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    const ro = new ResizeObserver(() => draw());
    ro.observe(container);
    return () => ro.disconnect();
  }, [draw]);

  return (
    <div
      ref={containerRef}
      className={className}
      data-graph-renderer="signalforge.tile.swimlane"
      data-graph-tile={tile.tile_id ?? tile.id}
      style={{ width: "100%", minWidth: 0, height: computedHeight, overflow: "hidden" }}
    >
      <canvas ref={canvasRef} style={{ display: "block" }} />
    </div>
  );
}

// Exported helpers so callers can reuse the default color logic.
export { defaultColorForValue as swimlaneDefaultColorForValue, defaultTextForFill as swimlaneDefaultTextForFill };
