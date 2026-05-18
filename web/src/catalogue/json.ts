import type { Channel, JsonSignalCatalogue, JsonSignalCatalogueEntry } from "../types";

export type JsonSignalCatalogueInput = Partial<Omit<JsonSignalCatalogue, "signals">> & {
  signals?: Array<Partial<JsonSignalCatalogueEntry>>;
  channels?: Array<Partial<Channel>>;
};

export function normalizeJsonSignalCatalogue(input: JsonSignalCatalogueInput): JsonSignalCatalogue {
  const signals = Array.isArray(input.signals) ? input.signals.map(normalizeSignalEntry).filter((signal): signal is JsonSignalCatalogueEntry => signal !== null) : [];
  const channels = Array.isArray(input.channels) ? input.channels.map(normalizeChannel).filter((channel): channel is Channel => channel !== null) : [];
  return {
    schema_version: input.schema_version,
    source: input.source,
    generated_at: input.generated_at,
    signals,
    channels,
    meta: isRecord(input.meta) ? input.meta : undefined,
  };
}

function normalizeSignalEntry(signal: Partial<JsonSignalCatalogueEntry>): JsonSignalCatalogueEntry | null {
  if (!signal || typeof signal.id !== "number" || !Number.isFinite(signal.id) || typeof signal.name !== "string" || !signal.name.trim()) return null;
  const role = signal.role === "control" ? "control" : "monitor";
  const kind = signal.kind === "int" || signal.kind === "enum" ? signal.kind : "float";
  return {
    id: signal.id,
    sid: optionalText(signal.sid),
    name: signal.name,
    group: text(signal.group, "default"),
    subgroup: text(signal.subgroup, "default"),
    role,
    kind,
    unit: optionalText(signal.unit),
    type: optionalText(signal.type),
    writable: signal.writable,
    dangerous: signal.dangerous,
    min: numberOrUndefined(signal.min),
    max: numberOrUndefined(signal.max),
    enum: isRecord(signal.enum) ? signal.enum as Record<string, string> : undefined,
    applicableModes: Array.isArray(signal.applicableModes) ? signal.applicableModes.filter((mode): mode is string => typeof mode === "string") : undefined,
    cmd: optionalText(signal.cmd),
    title: optionalText(signal.title),
    description: optionalText(signal.description),
    help: optionalText(signal.help),
    safety_note: optionalText(signal.safety_note),
    evidence_ref: optionalText(signal.evidence_ref),
    counterparts: Array.isArray(signal.counterparts) ? signal.counterparts.map((counterpart) => counterpart && typeof counterpart.id === "string" ? counterpart : null).filter((counterpart): counterpart is NonNullable<typeof counterpart> => counterpart !== null) : undefined,
  };
}

function normalizeChannel(channel: Partial<Channel>): Channel | null {
  if (!channel || typeof channel.device_id !== "string" || !channel.device_id.trim() || !Number.isFinite(Number(channel.instance))) return null;
  return {
    device_id: channel.device_id,
    instance: Number(channel.instance) || 1,
    role: text(channel.role, "unknown"),
    label: text(channel.label, channel.device_id),
    endpoint: optionalText(channel.endpoint),
  };
}

function text(value: unknown, fallback = "") {
  if (value === undefined || value === null || value === "") return fallback;
  return String(value);
}

function optionalText(value: unknown) {
  if (value === undefined || value === null || value === "") return undefined;
  return String(value);
}

function numberOrUndefined(value: unknown) {
  if (value === undefined || value === null || value === "") return undefined;
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
