import { normalizeJsonSignalCatalogue } from "../src/catalogue/json";

describe("normalizeJsonSignalCatalogue", () => {
  it("normalizes sparse catalogue JSON into the public catalogue shape", () => {
    const catalogue = normalizeJsonSignalCatalogue({
      schema_version: 1,
      source: "fixture",
      signals: [
        {
          id: 7,
          name: "Pressure",
          role: "monitor",
          kind: "float",
          group: "tvac",
          subgroup: "pressure",
          unit: "mbar",
          writable: false,
          dangerous: false,
          counterparts: [{ id: "pressure.actual", label: "Actual pressure" }],
        },
      ],
      channels: [
        {
          device_id: "tec-a",
          instance: "1",
          role: "monitor",
          label: "TEC A",
        },
      ],
    });

    expect(catalogue).toEqual({
      schema_version: 1,
      source: "fixture",
      generated_at: undefined,
      signals: [
        {
          id: 7,
          sid: undefined,
          name: "Pressure",
          group: "tvac",
          subgroup: "pressure",
          role: "monitor",
          kind: "float",
          unit: "mbar",
          type: undefined,
          writable: false,
          dangerous: false,
          min: undefined,
          max: undefined,
          enum: undefined,
          applicableModes: undefined,
          cmd: undefined,
          title: undefined,
          description: undefined,
          help: undefined,
          safety_note: undefined,
          evidence_ref: undefined,
          counterparts: [{ id: "pressure.actual", label: "Actual pressure" }],
        },
      ],
      channels: [
        {
          device_id: "tec-a",
          instance: 1,
          role: "monitor",
          label: "TEC A",
          endpoint: undefined,
        },
      ],
      meta: undefined,
    });
  });
});
