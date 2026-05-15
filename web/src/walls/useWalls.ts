import { useState, useEffect } from "react";
import type { WallConfig } from "../types";

function storageKey(namespace: string): string {
  return `${namespace}.walls`;
}

const changeEvent = (namespace: string) => `${namespace}-walls-changed`;

export function loadWalls(namespace: string): WallConfig[] {
  try {
    const raw = JSON.parse(localStorage.getItem(storageKey(namespace)) || "[]");
    if (!Array.isArray(raw)) return [];
    return raw.filter((w): w is WallConfig => w && typeof w.wall_id === "string" && typeof w.label === "string");
  } catch (_) { return []; }
}

export function saveWalls(list: WallConfig[], namespace: string): void {
  localStorage.setItem(storageKey(namespace), JSON.stringify(list));
  if (typeof window !== "undefined") {
    window.dispatchEvent(new CustomEvent(changeEvent(namespace)));
  }
}

export type WallsHandle = {
  walls: WallConfig[];
  add(label: string): WallConfig;
  rename(wallId: string, label: string): void;
  remove(wallId: string): void;
  wallForDevice(deviceId: string): WallConfig;
};

export function useWalls(namespace: string): WallsHandle {
  const [walls, setWalls] = useState<WallConfig[]>(() => loadWalls(namespace));
  useEffect(() => {
    const fn = () => setWalls(loadWalls(namespace));
    window.addEventListener(changeEvent(namespace), fn);
    return () => window.removeEventListener(changeEvent(namespace), fn);
  }, [namespace]);
  return {
    walls,
    add(label) {
      const wall: WallConfig = { wall_id: `${namespace}-wall-${Date.now()}`, label };
      saveWalls([...loadWalls(namespace), wall], namespace);
      return wall;
    },
    rename(wallId, label) {
      saveWalls(loadWalls(namespace).map((w) => w.wall_id === wallId ? { ...w, label } : w), namespace);
    },
    remove(wallId) {
      saveWalls(loadWalls(namespace).filter((w) => w.wall_id !== wallId), namespace);
    },
    wallForDevice(deviceId) {
      const id = `device-${deviceId}`;
      return { wall_id: id, label: `Device · ${deviceId}` };
    },
  };
}
