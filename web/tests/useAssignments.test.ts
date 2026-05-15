import { loadAssignments, saveAssignments, makeAssignment } from "../src/dict/useAssignments";
import type { AssignmentsStoreOptions } from "../src/types";

const store: AssignmentsStoreOptions = { namespace: "test-ns" };

// Minimal localStorage shim
const _store: Record<string, string> = {};
global.localStorage = {
  getItem: (k: string) => _store[k] ?? null,
  setItem: (k: string, v: string) => { _store[k] = v; },
  removeItem: (k: string) => { delete _store[k]; },
  clear: () => { Object.keys(_store).forEach((k) => delete _store[k]); },
  key: (i: number) => Object.keys(_store)[i] ?? null,
  get length() { return Object.keys(_store).length; },
} as Storage;


beforeEach(() => localStorage.clear());

describe("makeAssignment", () => {
  it("creates an assignment with defaults", () => {
    const a = makeAssignment("wall1", 42, "dev-A");
    expect(a.wall_id).toBe("wall1");
    expect(a.param_id).toBe(42);
    expect(a.device_id).toBe("dev-A");
    expect(a.instance).toBe(1);
    expect(a.kind).toBe("trend");
  });

  it("uses supplied instance", () => {
    const a = makeAssignment("wall1", 7, "dev-B", 3);
    expect(a.instance).toBe(3);
  });
});

describe("loadAssignments / saveAssignments", () => {
  it("returns empty array when nothing saved", () => {
    expect(loadAssignments(store)).toEqual([]);
  });

  it("round-trips correctly", () => {
    const a = makeAssignment("wall1", 1, "dev-A");
    saveAssignments([a], store);
    const loaded = loadAssignments(store);
    expect(loaded).toHaveLength(1);
    expect(loaded[0].param_id).toBe(1);
  });

  it("uses namespace key", () => {
    const a = makeAssignment("wall1", 2, "dev-A");
    saveAssignments([a], store);
    expect(localStorage.getItem("test-ns.assignments")).not.toBeNull();
    expect(localStorage.getItem("other.assignments")).toBeNull();
  });
});
