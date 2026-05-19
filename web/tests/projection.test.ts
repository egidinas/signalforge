import {
  buildSignalProjectionTree,
  CURRENT_SOURCE_PROJECTION_SCHEMA_VERSION,
  normalizeSignalProjectionBundle,
  signalProjectionPathKey,
  signalProjectionRefKey,
  validateSignalProjectionBundle,
} from "../src/catalogue/projection";

describe("signal projection contracts", () => {
  it("normalizes and validates MeCom-style primary and secondary mappings", () => {
    const bundle = normalizeSignalProjectionBundle({
      schema_version: CURRENT_SOURCE_PROJECTION_SCHEMA_VERSION,
      primary_mappings: [
        {
          bundle: "temperature-loop",
          ids: [1000, 3000],
          path: ["Temperature controllers", { id: "loop", label: "Control loop", order: 1 }],
          device_grouping: "device",
        },
      ],
      secondary_mappings: [
        {
          bundle: "quick-preview",
          ids: [1000],
          path: ["Quick preview"],
          reason: "Shown again as a compact recent-history preview.",
        },
      ],
    });

    expect(bundle.mappings).toHaveLength(2);
    expect(bundle.mappings[0].signal_refs.map(signalProjectionRefKey)).toEqual(["id:1000", "id:3000"]);
    expect(signalProjectionPathKey(bundle.mappings[0].path)).toBe("temperature-controllers/loop");
    expect(validateSignalProjectionBundle(bundle, { requiredPrimarySignalIds: [1000, 3000] })).toEqual({
      valid: true,
      errors: [],
      warnings: [],
    });
  });

  it("normalizes legacy meta input to canonical metadata output", () => {
    const bundle = normalizeSignalProjectionBundle({
      meta: { owner: "signalforge" },
      mappings: [
        {
          kind: "primary",
          ids: [1000],
          path: [{ id: "temperature", label: "Temperature", meta: { tooltip: "Measured temperature" } }],
          signal_refs: [{ signal_id: 1000, meta: { unit_hint: "Degree Celsius" } }],
          meta: { projection: "operator" },
        },
      ],
    });

    expect(bundle.metadata).toEqual({ owner: "signalforge" });
    expect(bundle.mappings[0].metadata).toEqual({ projection: "operator" });
    expect(bundle.mappings[0].path[0].metadata).toEqual({ tooltip: "Measured temperature" });
    expect(bundle.mappings[0].signal_refs[0].metadata).toEqual({ unit_hint: "Degree Celsius" });
    expect(JSON.stringify(bundle)).not.toContain("\"meta\"");
  });

  it("rejects stale projection schema versions", () => {
    const result = validateSignalProjectionBundle({
      schema_version: "signalforge.projection.v1",
      mappings: [
        { kind: "primary", path: [{ id: "temperature", label: "Temperature" }], signal_refs: [{ signal_id: 1000 }] },
      ],
    });

    expect(result.valid).toBe(false);
    expect(result.errors.map((error) => error.code)).toEqual(["schema_version"]);
  });

  it("rejects duplicate primary mappings and missing required primary mappings", () => {
    const bundle = normalizeSignalProjectionBundle({
      primary_mappings: [
        { bundle: "a", ids: [1000], path: ["A"] },
        { bundle: "b", ids: [1000], path: ["B"] },
      ],
    });

    const result = validateSignalProjectionBundle(bundle, { requiredPrimarySignalIds: [1000, 3000] });

    expect(result.valid).toBe(false);
    expect(result.errors.map((error) => error.code)).toEqual(["duplicate_primary", "missing_primary"]);
  });

  it("keeps OPC UA browse paths source-agnostic while preserving source refs", () => {
    const bundle = normalizeSignalProjectionBundle({
      mappings: [
        {
          kind: "primary",
          source_family: "opcua",
          source_id: "app03-opcua",
          trace_ids: ["ns=2;s=Hall.Environment.Temperature"],
          path: [
            { id: "opcua", label: "OPC UA", source_family: "opcua" },
            { id: "hall", label: "Hall environment" },
          ],
        },
      ],
    });

    expect(bundle.mappings[0].signal_refs).toEqual([
      {
        trace_id: "ns=2;s=Hall.Environment.Temperature",
        source_id: "app03-opcua",
        source_family: "opcua",
      },
    ]);
    expect(validateSignalProjectionBundle(bundle, {
      sourceCatalogues: [
        {
          source_id: "app03-opcua",
          source_family: "opcua",
          entries: [{ trace_id: "ns=2;s=Hall.Environment.Temperature" }],
        },
      ],
    }).valid).toBe(true);
  });

  it("supports DBC CAN message and signal grouping through the same tree", () => {
    const bundle = normalizeSignalProjectionBundle({
      mappings: [
        {
          kind: "primary",
          source_family: "can_dbc",
          source_id: "thermal-can",
          trace_ids: ["CondorMk3.0x124.OT_degC"],
          path: [
            { id: "can", label: "CAN DBC", source_family: "can_dbc" },
            { id: "condormk3", label: "CondorMk3" },
            { id: "temperature", label: "Temperature" },
          ],
        },
      ],
    });

    const tree = buildSignalProjectionTree(bundle);

    expect(tree).toHaveLength(1);
    expect(validateSignalProjectionBundle(bundle, {
      sourceCatalogues: [{
        source_id: "thermal-can",
        source_family: "can_dbc",
        entries: [{
          trace_id: "CondorMk3.0x124.OT_degC",
          group_key: "can_dbc:0x124:CondorMk3",
          group_label: "CondorMk3",
          instance_key: "0x124",
          sort_key: "00000124.000.OT_degC",
        }],
      }],
    }).valid).toBe(true);
    expect(tree[0].children[0].children[0].mappings[0].signal_refs.map(signalProjectionRefKey)).toEqual([
      "trace:can_dbc:thermal-can:CondorMk3.0x124.OT_degC",
    ]);
  });

  it("rejects source refs that are not present in supplied source catalogues", () => {
    const bundle = normalizeSignalProjectionBundle({
      mappings: [
        {
          kind: "primary",
          source_family: "can_dbc",
          source_id: "thermal-can",
          trace_ids: ["CondorMk3.0x124.missing"],
          path: [{ id: "temperature", label: "Temperature" }],
        },
      ],
    });

    const result = validateSignalProjectionBundle(bundle, {
      sourceCatalogues: [
        {
          source_id: "thermal-can",
          source_family: "can_dbc",
          entries: [{ trace_id: "CondorMk3.0x124.OT_degC" }],
        },
      ],
    });

    expect(result.valid).toBe(false);
    expect(result.errors.map((error) => error.code)).toEqual(["unknown_source_ref"]);
  });
});
