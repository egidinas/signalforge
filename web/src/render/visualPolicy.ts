import type { GraphTileCardRef, GraphWallCard, GraphWallModel, TileSeries } from "../types";

export interface VisualPolicyRule {
  readonly any?: readonly string[];
  readonly all?: readonly string[];
  readonly color: string;
}

export interface VisualPolicyConfig {
  readonly roleColors: Record<string, string>;
  readonly signalColors: Record<string, string>;
  readonly palette: readonly string[];
  readonly semanticColorRules: readonly VisualPolicyRule[];
  readonly eventColorRules: readonly VisualPolicyRule[];
  readonly defaultEventColor: string;
}

export type VisualPolicyConfigInput = Partial<VisualPolicyConfig>;

const DEFAULT_ROLE_COLORS: Record<string, string> = {
  command: "#ffd85f",
  ghost: "#8aa7c4",
  acceptance_band: "#3ddc84",
  actual: "#56d6df",
  dut: "#ff6b35",
  aux: "#9db4c8",
  source_quality: "#66b8ef",
  counter: "#b8a6ff",
  interlock: "#ff6374",
  evidence: "#b8a6ff",
  event: "#f2f7ff",
  state: "#8bd3a5",
};

const DEFAULT_SIGNAL_COLORS: Record<string, string> = {
  "trace.command.chamber": "#ffd400",
  "trace.ghost.profile": "#f8fafc",
  "trace.acceptance.temperature": "#4ee28a",
  "trace.actual.chamber_air": "#00c8ff",
  "trace.context.chamber_air": "#00c8ff",
  "trace.table_loop": "#ff8a00",
  "trace.interface": "#ff8a00",
  "trace.shroud": "#b65cff",
  "trace.shroud_inlet": "#00a8ff",
  "trace.shroud_outlet": "#ff58c8",
  "trace.dut_temp_a": "#ff315f",
  "trace.dut_temp_b": "#00d6a3",
  "trace.tvac_pressure": "#1f6fff",
  "trace.actual.tvac_pressure": "#1f6fff",
  "trace.tvac_pressure_target": "#9cc7ff",
  "trace.tvac_outgassing": "#ff5c93",
  "trace.tvac_virtual_leak": "#f5d742",
  "trace.tvac_roughing_pump": "#ff7a35",
  "trace.tvac_turbo_pump": "#31d6ff",
  "trace.tvac_pump_removal": "#b079ff",
  "trace.tvac_volatile_inventory": "#9bff70",
  "trace.total_power": "#f97316",
  "trace.subsystem_power": "#60a5fa",
  "trace.bus_packets": "#a78bfa",
  "trace.bus_retries": "#fb7185",
  "trace.phase_enum": "#e5e7eb",
  "trace.functional_gate_active": "#fbbf24",
  "trace.stability_reached": "#34d399",
  "trace.dwell_active": "#38bdf8",
  "trace.dwell_complete": "#a78bfa",
  "trace.dut_ready": "#84cc16",
  "trace.dut_operative": "#22c55e",
  "trace.payload_active": "#f97316",
  "trace.rf_link_locked": "#06b6d4",
  "trace.fault_flag": "#fb7185",
  "trace.power_total": "#ff7a00",
  "trace.power_subsystem": "#00c8ff",
  "trace.power_payload": "#ff315f",
  "trace.power_avionics": "#00d6a3",
  "trace.power_link": "#b079ff",
  "trace.overall_packet_counter": "#f8fafc",
  "trace.tm_packet_counter": "#00c8ff",
  "trace.tc_packet_counter": "#ffd400",
  "trace.dropped_frame_count": "#ff315f",
  "trace.bus_latency": "#ffb000",
  "trace.source_freshness": "#00d6a3",
  "trace.cooling_water_temp": "#00a8ff",
  "trace.pressurized_air_supply": "#ffd400",
  "trace.air_dewpoint": "#b079ff",
  "trace.ln2_duty": "#00a8ff",
  "trace.freeze_margin": "#4ee28a",
  "trace.tvac_cryo_exhaust": "#1f6fff",
  "trace.tvac_scavenged_exhaust": "#00d6a3",
  "trace.tvac_scavenger_water_return": "#ff8a00",
  "trace.tvac_exhaust_cold_recovery": "#b079ff",
};

const DEFAULT_DISTINCTIVE_PALETTE = [
  "#31d6ff",
  "#ffb000",
  "#ff5c93",
  "#00d084",
  "#b079ff",
  "#ff7a35",
  "#44e0b7",
  "#7aa2ff",
  "#f5d742",
  "#f15bb5",
  "#2ec4b6",
  "#e76f51",
  "#9bff70",
  "#00a6fb",
  "#ffd166",
  "#ef476f",
];

const DEFAULT_SEMANTIC_COLOR_RULES: VisualPolicyRule[] = [
  { any: ["dut_temp_a", "dut.a", "node_a"], color: "#ff315f" },
  { any: ["dut_temp_b", "dut.b", "node_b"], color: "#00d6a3" },
  { all: ["dut", "temp"], color: "#ff6b35" },
  { any: ["command", "target"], color: "#ffd400" },
  { any: ["ghost", "profile"], color: "#f8fafc" },
  { any: ["pressure"], color: "#1f6fff" },
  { any: ["power"], color: "#ff7a35" },
  { any: ["packet", "bus"], color: "#b079ff" },
  { any: ["ready", "operative", "stability"], color: "#00d6a3" },
  { any: ["fault", "error", "interlock"], color: "#ff315f" },
  { any: ["interface", "table", "platen"], color: "#ff8a00" },
  { any: ["shroud"], color: "#b65cff" },
  { any: ["chamber"], color: "#00c8ff" },
];

const DEFAULT_EVENT_COLOR_RULES: VisualPolicyRule[] = [
  { any: ["functional", "gate"], color: "#ffb000" },
  { any: ["evidence"], color: "#b079ff" },
  { any: ["interlock", "fault"], color: "#ff315f" },
  { any: ["stability", "dwell"], color: "#00d6a3" },
  { any: ["pressure"], color: "#1f6fff" },
];

export const DEFAULT_VISUAL_POLICY_CONFIG: VisualPolicyConfig = {
  roleColors: DEFAULT_ROLE_COLORS,
  signalColors: DEFAULT_SIGNAL_COLORS,
  palette: DEFAULT_DISTINCTIVE_PALETTE,
  semanticColorRules: DEFAULT_SEMANTIC_COLOR_RULES,
  eventColorRules: DEFAULT_EVENT_COLOR_RULES,
  defaultEventColor: "#31d6ff",
};

export const roleColors: Record<string, string> = { ...DEFAULT_ROLE_COLORS };
export const signalColors: Record<string, string> = { ...DEFAULT_SIGNAL_COLORS };
export const distinctivePalette = [...DEFAULT_DISTINCTIVE_PALETTE];

let semanticColorRules = [...DEFAULT_SEMANTIC_COLOR_RULES];
let eventColorRules = [...DEFAULT_EVENT_COLOR_RULES];
let defaultEventColor = DEFAULT_VISUAL_POLICY_CONFIG.defaultEventColor;

function replaceRecord(target: Record<string, string>, next: Record<string, string>) {
  Object.keys(target).forEach((key) => delete target[key]);
  Object.assign(target, next);
}

function replaceArray<T>(target: T[], next: readonly T[]) {
  target.splice(0, target.length, ...next);
}

function cloneRules(rules: readonly VisualPolicyRule[]) {
  return rules.map((rule) => ({
    ...rule,
    any: rule.any ? [...rule.any] : undefined,
    all: rule.all ? [...rule.all] : undefined,
  }));
}

export function createVisualPolicyConfig(input: VisualPolicyConfigInput = {}): VisualPolicyConfig {
  return {
    roleColors: { ...DEFAULT_ROLE_COLORS, ...(input.roleColors ?? {}) },
    signalColors: { ...DEFAULT_SIGNAL_COLORS, ...(input.signalColors ?? {}) },
    palette: input.palette && input.palette.length > 0 ? [...input.palette] : [...DEFAULT_DISTINCTIVE_PALETTE],
    semanticColorRules: input.semanticColorRules ? cloneRules(input.semanticColorRules) : cloneRules(DEFAULT_SEMANTIC_COLOR_RULES),
    eventColorRules: input.eventColorRules ? cloneRules(input.eventColorRules) : cloneRules(DEFAULT_EVENT_COLOR_RULES),
    defaultEventColor: input.defaultEventColor ?? DEFAULT_VISUAL_POLICY_CONFIG.defaultEventColor,
  };
}

export function configureVisualPolicy(input: VisualPolicyConfigInput = {}) {
  const next = createVisualPolicyConfig(input);
  replaceRecord(roleColors, next.roleColors);
  replaceRecord(signalColors, next.signalColors);
  replaceArray(distinctivePalette, next.palette);
  semanticColorRules = cloneRules(next.semanticColorRules);
  eventColorRules = cloneRules(next.eventColorRules);
  defaultEventColor = next.defaultEventColor;
  return getVisualPolicyConfig();
}

export function resetVisualPolicy() {
  return configureVisualPolicy();
}

export function getVisualPolicyConfig(): VisualPolicyConfig {
  return {
    roleColors: { ...roleColors },
    signalColors: { ...signalColors },
    palette: [...distinctivePalette],
    semanticColorRules: cloneRules(semanticColorRules),
    eventColorRules: cloneRules(eventColorRules),
    defaultEventColor,
  };
}

export function palette(index: number) {
  return distinctivePalette[index % distinctivePalette.length];
}

export function paletteForID(id: string, fallbackIndex: number) {
  let hash = fallbackIndex + 17;
  for (let i = 0; i < id.length; i += 1) hash = ((hash << 5) - hash + id.charCodeAt(i)) | 0;
  return distinctivePalette[Math.abs(hash) % distinctivePalette.length];
}

export function colorForSignal(signal: Pick<TileSeries, "id" | "role" | "render_kind" | "kind" | "color"> | { id: string; role: string; kind?: string; color?: string }, index = 0) {
  const kind = "kind" in signal ? signal.kind : ("render_kind" in signal ? signal.render_kind : undefined);
  const configuredColor = "color" in signal ? signal.color : undefined;
  if (typeof configuredColor === "string" && configuredColor.trim()) return configuredColor.trim();
  if (signalColors[signal.id]) return signalColors[signal.id];
  const semantic = semanticColor(signal.id);
  if (semantic) return semantic;
  const roleColor = roleColors[signal.role];
  if (roleColor) return roleColor;
  const kindColor = kind ? roleColors[kind] : undefined;
  if (kindColor) return kindColor;
  return paletteForID(signal.id, index) ?? palette(index);
}

function ruleMatches(lower: string, rule: VisualPolicyRule) {
  const all = rule.all ?? [];
  const any = rule.any ?? [];
  return all.every((term) => lower.includes(term.toLowerCase()))
    && (any.length === 0 || any.some((term) => lower.includes(term.toLowerCase())));
}

function colorForRuleText(text: string, rules: readonly VisualPolicyRule[]) {
  const lower = text.toLowerCase();
  return rules.find((rule) => ruleMatches(lower, rule))?.color;
}

export function semanticColor(id: string) {
  return colorForRuleText(id, semanticColorRules);
}

export function signalPriority(signal: { id: string; label?: string; role?: string; kind?: string; render_kind?: string }) {
  const text = `${signal.id} ${signal.label ?? ""}`.toLowerCase();
  if (signal.role === "command") return 0;
  if (signal.role === "ghost") return 1;
  if (signal.role === "acceptance_band") return 2;
  if (text.includes("dut")) return 3;
  if (text.includes("article") || text.includes("component")) return 4;
  if (text.includes("interface") || text.includes("platen") || text.includes("table")) return 5;
  if (text.includes("chamber") || text.includes("shroud")) return 6;
  if (text.includes("pressure")) return 7;
  if (text.includes("power")) return 8;
  if (text.includes("bus") || text.includes("packet")) return 9;
  if (signal.kind === "state" || signal.render_kind === "swimlane") return 10;
  return 20;
}

export function orderLegendSignals<T extends { id: string; label?: string; role?: string; kind?: string; render_kind?: string }>(signals: T[]) {
  return [...signals].sort((a, b) => signalPriority(a) - signalPriority(b));
}

export function graphCardPriority(a: GraphWallCard, b: GraphWallCard) {
  return graphCardRank(a) - graphCardRank(b);
}

export function graphSectionPriority(a: GraphWallModel["sections"][number], b: GraphWallModel["sections"][number]) {
  return graphSectionRank(a) - graphSectionRank(b);
}

export function graphSectionRank(section: GraphWallModel["sections"][number]) {
  return Math.min(...section.cards.map(graphCardRank), 100);
}

export function graphCardRank(card: GraphWallCard) {
  const id = card.id.toLowerCase();
  const title = card.title.toLowerCase();
  if (id === "thermal_program") return 0;
  if (id.includes("dut_temperature") || title.includes("dut temperature")) return 10;
  if (id.includes("dut_power") || title.includes("dut power")) return 20;
  if (id.includes("tmtc_health")) return 30;
  if (id.includes("tmtc_counters")) return 40;
  if (id.includes("state_change") || card.render_kind === "swimlane") return 50;
  if (id.includes("functional_events") || card.render_kind === "event_rail") return 60;
  if (id.includes("facility") || id.includes("building") || id.includes("source_quality") || title.includes("testbed")) return 80;
  return 70;
}

export function cardPriority(card: GraphTileCardRef) {
  const order: Record<string, number> = {
    thermal_program: 0,
    dut_temperature: 1,
    tvac_pressure: 2,
    facility_actuation: 3,
    dut_power: 4,
    tmtc_counters: 5,
    state_change_swimlane: 6,
    functional_events: 7,
    tvac_exhaust_scavenger: 8,
    building_infrastructure: 8,
    facility_temperature_safety: 9,
    tmtc_health: 10,
    source_quality: 11,
  };
  return order[card.card_id] ?? 40;
}

export function tileCardPriority(a: GraphTileCardRef, b: GraphTileCardRef) {
  const aPriority = cardPriority(a);
  const bPriority = cardPriority(b);
  if (aPriority !== bPriority) return aPriority - bPriority;
  if (a.default_expanded !== b.default_expanded) return a.default_expanded ? -1 : 1;
  return a.card_id.localeCompare(b.card_id);
}

export function eventColor(kind?: string) {
  return colorForRuleText(kind ?? "", eventColorRules) ?? defaultEventColor;
}

export function blockLabel(label: string, value: number) {
  const normalized = String(label ?? "").trim();
  if (normalized && normalized !== "0" && normalized !== "1") return normalized;
  return value > 0 ? "ACTIVE" : "idle";
}
