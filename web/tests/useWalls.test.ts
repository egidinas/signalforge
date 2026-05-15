import { loadWalls, saveWalls } from "../src/walls/useWalls";

const NS = "test-walls";

const _store: Record<string, string> = {};
global.localStorage = {
  getItem: (k: string) => _store[k] ?? null,
  setItem: (k: string, v: string) => { _store[k] = v; },
  removeItem: (k: string) => { delete _store[k]; },
  clear: () => { Object.keys(_store).forEach((k) => delete _store[k]); },
  key: (i: number) => Object.keys(_store)[i] ?? null,
  get length() { return Object.keys(_store).length; },
} as Storage;

global.dispatchEvent = () => true;

beforeEach(() => localStorage.clear());

describe("loadWalls / saveWalls", () => {
  it("returns empty list when nothing saved", () => {
    expect(loadWalls(NS)).toEqual([]);
  });

  it("round-trips a wall", () => {
    const w = { wall_id: "w-1", label: "Alpha" };
    saveWalls([w], NS);
    const list = loadWalls(NS);
    expect(list).toHaveLength(1);
    expect(list[0].label).toBe("Alpha");
  });

  it("persists under namespace key", () => {
    saveWalls([{ wall_id: "w-2", label: "Beta" }], NS);
    expect(localStorage.getItem(`${NS}.walls`)).not.toBeNull();
    expect(localStorage.getItem("other.walls")).toBeNull();
  });

  it("preserves preset flag", () => {
    saveWalls([{ wall_id: "w-3", label: "Preset Wall", preset: true }], NS);
    const list = loadWalls(NS);
    expect(list[0].preset).toBe(true);
  });
});
