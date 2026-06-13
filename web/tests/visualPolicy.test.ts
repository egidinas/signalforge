import {
  colorForSignal,
  configureVisualPolicy,
  getVisualPolicyConfig,
  resetVisualPolicy,
  roleColors,
} from "../src/render/visualPolicy";

describe("visualPolicy", () => {
  afterEach(() => {
    resetVisualPolicy();
  });

  it("uses role colors before hashed palette fallbacks", () => {
    expect(colorForSignal({ id: "opaque-dut-id", role: "dut" })).toBe(roleColors.dut);
    expect(colorForSignal({ id: "opaque-aux-id", role: "aux" })).toBe(roleColors.aux);
    expect(colorForSignal({ id: "opaque-state-id", role: "state" })).toBe(roleColors.state);
  });

  it("keeps graph colors configuration driven", () => {
    configureVisualPolicy({
      roleColors: { dut: "#101820" },
      signalColors: { "trace.configured": "var(--sf-series-configured)" },
      palette: ["#123456"],
    });

    expect(getVisualPolicyConfig().roleColors.dut).toBe("#101820");
    expect(roleColors.dut).toBe("#101820");
    expect(colorForSignal({ id: "opaque-dut-id", role: "dut" })).toBe("#101820");
    expect(colorForSignal({ id: "trace.configured", role: "actual" })).toBe("var(--sf-series-configured)");
    expect(colorForSignal({ id: "unknown-signal", role: "unknown" })).toBe("#123456");
  });

  it("honors contract-provided CSS variable colors", () => {
    expect(colorForSignal({ id: "trace.css-var", role: "actual", color: "var(--series-actual)" })).toBe("var(--series-actual)");
  });
});
