import { colorForSignal, roleColors } from "../src/render/visualPolicy";

describe("visualPolicy", () => {
  it("uses role colors before hashed palette fallbacks", () => {
    expect(colorForSignal({ id: "opaque-dut-id", role: "dut" })).toBe(roleColors.dut);
    expect(colorForSignal({ id: "opaque-aux-id", role: "aux" })).toBe(roleColors.aux);
    expect(colorForSignal({ id: "opaque-state-id", role: "state" })).toBe(roleColors.state);
  });
});
