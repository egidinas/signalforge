import React from "react";

export type SparklineProps = {
  data: number[];
  width?: number;
  height?: number;
  color?: string;
  strokeWidth?: number;
  min?: number;
  max?: number;
};

export function Sparkline({
  data,
  width = 60,
  height = 20,
  color = "var(--brand)",
  strokeWidth = 1.5,
  min: forceMin,
  max: forceMax,
}: SparklineProps) {
  if (!data || data.length < 2) return null;

  const min = forceMin !== undefined ? forceMin : Math.min(...data);
  const max = forceMax !== undefined ? forceMax : Math.max(...data);
  const range = max - min;
  const step = width / (data.length - 1);

  const points = data.map((v, i) => {
    const x = i * step;
    const y = range === 0 ? height / 2 : height - ((v - min) / range) * height;
    return `${x.toFixed(2)},${y.toFixed(2)}`;
  }).join(" ");

  return (
    <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`} style={{ overflow: "visible" }}>
      <polyline
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeLinejoin="round"
        points={points}
      />
    </svg>
  );
}
