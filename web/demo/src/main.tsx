import React, { useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import "uplot/dist/uPlot.min.css";
import "../src/styles.css";
import {
  SignalDictionary,
  UPlotTileRenderer,
  WallManager,
  makeAssignment,
  useAssignments,
  useWalls,
} from "../../src";
import type {
  Assignment,
  Channel,
  GraphTile,
  SemanticSignal,
  SignalCatalogueAdapter,
} from "../../src";

const namespace = "signalforge-web-demo";
const demoWallId = "demo-wall-primary";

const signals: SemanticSignal[] = [
  { id: 101, name: "Chamber temperature", group: "Thermal", subgroup: "Environment", role: "monitor", kind: "float", unit: "degC" },
  { id: 102, name: "Target temperature", group: "Thermal", subgroup: "Command", role: "monitor", kind: "float", unit: "degC" },
  { id: 201, name: "Bus voltage", group: "Power", subgroup: "Supply", role: "monitor", kind: "float", unit: "V" },
  { id: 202, name: "Bus current", group: "Power", subgroup: "Supply", role: "monitor", kind: "float", unit: "A" },
  { id: 301, name: "Frame counter", group: "Telemetry", subgroup: "Transport", role: "monitor", kind: "int" },
  { id: 401, name: "Set thermal target", group: "Thermal", subgroup: "Command", role: "control", kind: "float", unit: "degC", writable: true },
  { id: 402, name: "Enable profile hold", group: "Thermal", subgroup: "Command", role: "control", kind: "enum", writable: true },
];

const channels: Channel[] = [
  { device_id: "sim-a", instance: 1, role: "primary", label: "Simulator A" },
  { device_id: "sim-b", instance: 1, role: "secondary", label: "Simulator B" },
];

const adapter: SignalCatalogueAdapter = {
  list: () => signals,
  channels: () => channels,
  channelsForSignal: (signal) => signal.role === "monitor" ? channels : [],
  subscribeLive: () => () => undefined,
  formatValue: (value, unit) => value == null ? "n/a" : `${value.toFixed(2)}${unit ? ` ${unit}` : ""}`,
  roleForParam: (paramId) => signals.find((signal) => signal.id === paramId)?.role ?? "monitor",
  colorForRole: (role) => role === "control" ? "#f59e0b" : "#22d3ee",
};

function App() {
  const walls = useWalls(namespace);
  const assignments = useAssignments({ namespace });
  const [selectedWallId, setSelectedWallId] = useState(demoWallId);

  const wallList = useMemo(() => {
    const preset = { wall_id: demoWallId, label: "Synthetic system", preset: true };
    return walls.walls.some((wall) => wall.wall_id === demoWallId)
      ? walls.walls
      : [preset, ...walls.walls];
  }, [walls.walls]);

  const selectedAssignments = assignments.forWall(selectedWallId);
  const plottedAssignments = selectedAssignments.length
    ? selectedAssignments
    : [
      makeAssignment(demoWallId, 101, "sim-a"),
      makeAssignment(demoWallId, 102, "sim-a"),
      makeAssignment(demoWallId, 201, "sim-a"),
    ];
  const tile = useMemo(() => buildTile(plottedAssignments), [plottedAssignments]);

  return (
    <main className="demo-shell">
      <section className="demo-toolbar">
        <div>
          <h1>SignalForge Web</h1>
          <p>Reusable public UI primitives with synthetic telemetry fixtures.</p>
        </div>
        <div className="demo-stats">
          <span>{wallList.length} walls</span>
          <span>{selectedAssignments.length} assigned</span>
          <span>{tile.series.length} plotted</span>
        </div>
      </section>

      <section className="demo-grid">
        <aside className="demo-side">
          <h2>Walls</h2>
          <WallManager
            walls={{ ...walls, walls: wallList }}
            selectedWallId={selectedWallId}
            onSelect={setSelectedWallId}
          />
        </aside>

        <section className="demo-main">
          <div className="demo-panel">
            <div className="demo-panel-head">
              <h2>Tile Renderer</h2>
              <span>graph_tile.v1 fixture</span>
            </div>
            <UPlotTileRenderer tile={tile} height={330} />
          </div>
          <div className="demo-panel">
            <div className="demo-panel-head">
              <h2>Signal Dictionary</h2>
              <span>local assignment store</span>
            </div>
            <SignalDictionary adapter={adapter} store={{ namespace }} wallId={selectedWallId} />
          </div>
        </section>
      </section>
    </main>
  );
}

function buildTile(assignments: Assignment[]): GraphTile {
  const end = Date.now();
  const start = end - 30 * 60 * 1000;
  const points = Array.from({ length: 91 }, (_, index) => start + index * 20_000);
  const series = assignments.map((assignment, index) => {
    const signal = signals.find((item) => item.id === assignment.param_id);
    return {
      id: `trace.${assignment.device_id}.${assignment.param_id}.${assignment.instance}`,
      label: `${signal?.name ?? assignment.param_id} / ${assignment.device_id}`,
      role: assignment.param_id === 102 ? "command" : "actual",
      unit: signal?.unit,
      units: signal?.unit,
      source: "synthetic-demo",
      source_family: "simulated",
      points: points.map((timestamp, pointIndex) => ({
        timestamp: new Date(timestamp).toISOString(),
        value: demoValue(assignment.param_id, pointIndex, index),
      })),
    };
  });

  return {
    schema_version: 1,
    id: "demo.synthetic.tile",
    card_id: "demo-card",
    level: "live",
    t0: new Date(start).toISOString(),
    t1: new Date(end).toISOString(),
    series,
    bands: [{
      id: "hold-window",
      kind: "dwell",
      label: "hold",
      start: new Date(start + 11 * 60 * 1000).toISOString(),
      end: new Date(start + 20 * 60 * 1000).toISOString(),
    }],
    markers: [{
      id: "profile-start",
      label: "profile",
      timestamp: new Date(start + 7 * 60 * 1000).toISOString(),
      kind: "operator",
      result: "accepted",
    }],
    events: [],
    diagnostics: {
      raw_point_count: points.length * Math.max(series.length, 1),
      point_count: points.length * Math.max(series.length, 1),
      decimation: "none",
      freshness_ms: 0,
      source: "synthetic-demo",
      mode: "demo",
    },
    provenance: {
      source: "signalforge-web-demo",
      source_family: "simulated",
      generated_at: new Date().toISOString(),
      synthetic: true,
    },
  };
}

function demoValue(paramId: number, pointIndex: number, seriesIndex: number) {
  const wave = Math.sin((pointIndex + seriesIndex * 4) / 8);
  if (paramId === 101) return 24 + wave * 5 + pointIndex * 0.03;
  if (paramId === 102) return 28 + Math.sin(pointIndex / 18) * 2;
  if (paramId === 201) return 28 + wave * 0.7;
  if (paramId === 202) return 2.6 + wave * 0.25;
  if (paramId === 301) return pointIndex;
  return 10 + wave * 3;
}

createRoot(document.getElementById("root") as HTMLElement).render(<App />);
