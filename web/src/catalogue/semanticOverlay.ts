import { useEffect, useState } from "react";
import type {
  SemanticOverlayBundle,
  SemanticOverlayEntry,
  SemanticOverlayStoreOptions,
  SemanticOverlayTarget,
} from "../types";

const DEFAULT_SCHEMA_VERSION = "semantic_overlay.v1";

function store(opts: SemanticOverlayStoreOptions): Storage | null {
  if (opts.storage) return opts.storage;
  if (typeof localStorage !== "undefined") return localStorage;
  return null;
}

function storageKey(namespace: string): string {
  return `${namespace}.semanticOverlay`;
}

function cleanString(value: unknown): string | undefined {
  const text = String(value ?? "").trim();
  return text || undefined;
}

function cleanNumberString(value: unknown): string | undefined {
  if (value === undefined || value === null || value === "") return undefined;
  const numeric = Number(value);
  return Number.isFinite(numeric) ? String(numeric) : cleanString(value);
}

export function semanticOverlayTargetKey(target: Partial<SemanticOverlayTarget> | undefined): string {
  const t = target || {};
  const direct = cleanString(t.target_id);
  if (direct) return direct;
  return [
    cleanString(t.kind) || "target",
    cleanString(t.device_id) || cleanString(t.serial) || "_",
    cleanNumberString(t.channel) || cleanNumberString(t.instance) || "_",
    cleanNumberString(t.signal_id) || "_",
    cleanString(t.group) || "_",
  ].join(":");
}

export function channelSemanticTarget(channel: { device_id?: string; serial?: string; instance?: number | string; role?: string }): SemanticOverlayTarget {
  return {
    kind: "channel",
    device_id: channel.device_id,
    serial: channel.serial,
    channel: channel.instance,
    instance: channel.instance,
    group: channel.role,
  };
}

export function normalizeSemanticOverlayEntry(input: Partial<SemanticOverlayEntry>): SemanticOverlayEntry | null {
  const target = input && input.target ? input.target : {};
  const key = semanticOverlayTargetKey(target);
  if (!key || key.includes(":_:_:_:_")) return null;
  const alias = cleanString(input.alias);
  const label = cleanString(input.label);
  const note = cleanString(input.note);
  const fixtureNote = cleanString(input.fixture_note);
  const source = cleanString(input.source);
  const author = cleanString(input.author);
  const updatedAt = cleanString(input.updated_at);
  const tags = Array.isArray(input.tags) ? input.tags.map(cleanString).filter((tag): tag is string => Boolean(tag)) : undefined;
  const meta = input.meta && typeof input.meta === "object" && !Array.isArray(input.meta) ? input.meta : undefined;
  const normalizedTarget: SemanticOverlayTarget = {
    target_id: cleanString(target.target_id),
    device_id: cleanString(target.device_id),
    serial: cleanString(target.serial),
    channel: cleanNumberString(target.channel),
    instance: cleanNumberString(target.instance),
    signal_id: cleanNumberString(target.signal_id),
    kind: cleanString(target.kind),
    group: cleanString(target.group),
  };
  Object.keys(normalizedTarget).forEach((field) => {
    if ((normalizedTarget as Record<string, unknown>)[field] === undefined) delete (normalizedTarget as Record<string, unknown>)[field];
  });
  return {
    id: cleanString(input.id) || key,
    target: normalizedTarget,
    ...(alias ? { alias } : {}),
    ...(label ? { label } : {}),
    ...(note ? { note } : {}),
    ...(fixtureNote ? { fixture_note: fixtureNote } : {}),
    ...(input.hidden !== undefined ? { hidden: Boolean(input.hidden) } : {}),
    ...(tags && tags.length ? { tags } : {}),
    ...(source ? { source } : {}),
    ...(author ? { author } : {}),
    ...(updatedAt ? { updated_at: updatedAt } : {}),
    ...(meta ? { meta } : {}),
  };
}

export function normalizeSemanticOverlayBundle(input: Partial<SemanticOverlayBundle> | SemanticOverlayEntry[] | null | undefined, namespace = "signalforge"): SemanticOverlayBundle {
  const bundle = Array.isArray(input) ? { entries: input } : (input || {});
  const entries = Array.isArray(bundle.entries)
    ? bundle.entries.map((entry) => normalizeSemanticOverlayEntry(entry)).filter((entry): entry is SemanticOverlayEntry => entry !== null)
    : [];
  const deduped = new Map<string, SemanticOverlayEntry>();
  entries.forEach((entry) => deduped.set(semanticOverlayTargetKey(entry.target), entry));
  return {
    schema_version: bundle.schema_version || DEFAULT_SCHEMA_VERSION,
    namespace: cleanString(bundle.namespace) || namespace,
    generated_at: cleanString(bundle.generated_at),
    entries: Array.from(deduped.values()).sort((a, b) => semanticOverlayTargetKey(a.target).localeCompare(semanticOverlayTargetKey(b.target))),
    ...(bundle.meta && typeof bundle.meta === "object" && !Array.isArray(bundle.meta) ? { meta: bundle.meta } : {}),
  };
}

export function overlayEntryForTarget(bundle: Partial<SemanticOverlayBundle> | SemanticOverlayEntry[] | null | undefined, target: Partial<SemanticOverlayTarget>): SemanticOverlayEntry | null {
  const key = semanticOverlayTargetKey(target);
  const normalized = normalizeSemanticOverlayBundle(bundle);
  return normalized.entries.find((entry) => semanticOverlayTargetKey(entry.target) === key) || null;
}

export function upsertSemanticOverlayEntry(bundle: Partial<SemanticOverlayBundle> | SemanticOverlayEntry[] | null | undefined, entry: Partial<SemanticOverlayEntry>, namespace = "signalforge"): SemanticOverlayBundle {
  const normalized = normalizeSemanticOverlayBundle(bundle, namespace);
  const next = normalizeSemanticOverlayEntry({
    ...entry,
    updated_at: entry.updated_at || new Date().toISOString(),
  });
  if (!next) return normalized;
  const key = semanticOverlayTargetKey(next.target);
  return normalizeSemanticOverlayBundle({
    ...normalized,
    generated_at: new Date().toISOString(),
    entries: normalized.entries.filter((item) => semanticOverlayTargetKey(item.target) !== key).concat(next),
  }, normalized.namespace || namespace);
}

export function removeSemanticOverlayEntry(bundle: Partial<SemanticOverlayBundle> | SemanticOverlayEntry[] | null | undefined, target: Partial<SemanticOverlayTarget>, namespace = "signalforge"): SemanticOverlayBundle {
  const normalized = normalizeSemanticOverlayBundle(bundle, namespace);
  const key = semanticOverlayTargetKey(target);
  return normalizeSemanticOverlayBundle({
    ...normalized,
    generated_at: new Date().toISOString(),
    entries: normalized.entries.filter((entry) => semanticOverlayTargetKey(entry.target) !== key),
  }, normalized.namespace || namespace);
}

export function loadSemanticOverlay(opts: SemanticOverlayStoreOptions): SemanticOverlayBundle {
  try {
    const raw = store(opts)?.getItem(storageKey(opts.namespace));
    return normalizeSemanticOverlayBundle(raw ? JSON.parse(raw) : null, opts.namespace);
  } catch (_) {
    return normalizeSemanticOverlayBundle(null, opts.namespace);
  }
}

export function saveSemanticOverlay(bundle: Partial<SemanticOverlayBundle> | SemanticOverlayEntry[], opts: SemanticOverlayStoreOptions): SemanticOverlayBundle {
  const normalized = normalizeSemanticOverlayBundle(bundle, opts.namespace);
  store(opts)?.setItem(storageKey(opts.namespace), JSON.stringify(normalized, null, 2));
  if (typeof window !== "undefined") {
    window.dispatchEvent(new CustomEvent(`${opts.namespace}-semantic-overlay-changed`));
  }
  return normalized;
}

export type SemanticOverlayHandle = {
  bundle: SemanticOverlayBundle;
  entryFor(target: Partial<SemanticOverlayTarget>): SemanticOverlayEntry | null;
  upsert(entry: Partial<SemanticOverlayEntry>): SemanticOverlayBundle;
  remove(target: Partial<SemanticOverlayTarget>): SemanticOverlayBundle;
};

export function useSemanticOverlay(opts: SemanticOverlayStoreOptions): SemanticOverlayHandle {
  const [bundle, setBundle] = useState<SemanticOverlayBundle>(() => loadSemanticOverlay(opts));
  useEffect(() => {
    const key = `${opts.namespace}-semantic-overlay-changed`;
    const fn = () => setBundle(loadSemanticOverlay(opts));
    window.addEventListener(key, fn);
    return () => window.removeEventListener(key, fn);
  }, [opts.namespace]);
  return {
    bundle,
    entryFor(target) {
      return overlayEntryForTarget(bundle, target);
    },
    upsert(entry) {
      const next = upsertSemanticOverlayEntry(bundle, entry, opts.namespace);
      saveSemanticOverlay(next, opts);
      return next;
    },
    remove(target) {
      const next = removeSemanticOverlayEntry(bundle, target, opts.namespace);
      saveSemanticOverlay(next, opts);
      return next;
    },
  };
}
