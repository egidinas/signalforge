import { useState, useEffect } from "react";
import type { Assignment, AssignmentsStoreOptions } from "../types";

function storageKey(namespace: string): string {
  return `${namespace}.assignments`;
}

function signalAddress(paramId: number, deviceId: string, instance: number): string {
  return `${paramId}@${deviceId}/${instance}`;
}

function parseSignalAddress(targetId: string): { param_id: number; device_id: string; instance: number } | null {
  const m = String(targetId || "").match(/^(\d+)@([^/]+)\/(\d+)$/);
  if (!m) return null;
  return { param_id: parseInt(m[1], 10), device_id: m[2], instance: parseInt(m[3], 10) || 1 };
}

function normalizeAssignment(a: Partial<Assignment>, namespace: string): Assignment | null {
  const item = a || {};
  const parsed = parseSignalAddress((item as Record<string, unknown>).target_id as string ?? "");
  const opts = (item.options as Record<string, unknown>) || {};
  const paramId = Number((item.param_id ?? (opts.param_id as number) ?? parsed?.param_id) ?? NaN);
  const deviceId = String((item.device_id ?? (opts.device_id as string) ?? parsed?.device_id) ?? "");
  const instance = Number((item.instance ?? (opts.instance as number) ?? parsed?.instance) ?? 1) || 1;
  const wallId = String((item.wall_id ?? "") || "wall");
  if (!deviceId || !Number.isFinite(paramId)) return null;
  const targetId = (item.target_id as string) ?? signalAddress(paramId, deviceId, instance);
  const tileId = (item.tile_id as string) ?? `${wallId}-${targetId}`;
  return {
    wall_id: wallId,
    tile_id: tileId,
    target_id: targetId,
    kind: (item.kind as string) ?? "trend",
    options: { ...opts, param_id: paramId, device_id: deviceId, instance },
    param_id: paramId,
    device_id: deviceId,
    instance,
  };
}

export function loadAssignments(opts: AssignmentsStoreOptions): Assignment[] {
  try {
    const raw = JSON.parse(localStorage.getItem(storageKey(opts.namespace)) || "[]");
    if (!Array.isArray(raw)) return [];
    return raw.map((a) => normalizeAssignment(a, opts.namespace)).filter((a): a is Assignment => a !== null);
  } catch (_) { return []; }
}

export function saveAssignments(list: Assignment[], opts: AssignmentsStoreOptions): void {
  const normalized = list.map((a) => normalizeAssignment(a, opts.namespace)).filter((a): a is Assignment => a !== null);
  localStorage.setItem(storageKey(opts.namespace), JSON.stringify(normalized));
  if (typeof window !== "undefined") {
    window.dispatchEvent(new CustomEvent(`${opts.namespace}-assignments-changed`));
  }
}

export function makeAssignment(wallId: string, paramId: number, deviceId: string, instance = 1): Assignment {
  const targetId = signalAddress(paramId, deviceId, instance);
  return {
    wall_id: wallId,
    tile_id: `${wallId}-${targetId}`,
    target_id: targetId,
    kind: "trend",
    options: { param_id: paramId, device_id: deviceId, instance },
    param_id: paramId,
    device_id: deviceId,
    instance,
  };
}

export type AssignmentsHandle = {
  list: Assignment[];
  add(wallId: string, paramId: number, deviceId: string, instance?: number): void;
  remove(wallId: string, paramId: number, deviceId: string, instance?: number): void;
  forWall(wallId: string): Assignment[];
  hasAssignment(wallId: string, paramId: number, deviceId: string, instance?: number): boolean;
};

export function useAssignments(opts: AssignmentsStoreOptions): AssignmentsHandle {
  const [list, setList] = useState<Assignment[]>(() => loadAssignments(opts));
  useEffect(() => {
    const key = `${opts.namespace}-assignments-changed`;
    const fn = () => setList(loadAssignments(opts));
    window.addEventListener(key, fn);
    return () => window.removeEventListener(key, fn);
  }, [opts.namespace]);
  return {
    list,
    add(wallId, paramId, deviceId, instance = 1) {
      const cur = loadAssignments(opts);
      const next = makeAssignment(wallId, paramId, deviceId, instance);
      if (cur.find((a) => a.wall_id === wallId && a.target_id === next.target_id)) return;
      saveAssignments([...cur, next], opts);
    },
    remove(wallId, paramId, deviceId, instance = 1) {
      const addr = signalAddress(paramId, deviceId, instance);
      saveAssignments(loadAssignments(opts).filter((a) => !(a.wall_id === wallId && a.target_id === addr)), opts);
    },
    forWall(wallId) {
      return list.filter((a) => a.wall_id === wallId);
    },
    hasAssignment(wallId, paramId, deviceId, instance = 1) {
      const addr = signalAddress(paramId, deviceId, instance);
      return list.some((a) => a.wall_id === wallId && a.target_id === addr);
    },
  };
}
