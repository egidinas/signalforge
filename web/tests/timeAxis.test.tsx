import { clampRange } from "../src/render/timeAxis";

describe("timeAxis", () => {
  it("does not expand a clamped range beyond the available full span", () => {
    expect(clampRange({ start: 25, end: 50 }, { start: 0, end: 100 }, 1_000)).toEqual({ start: 0, end: 100 });
  });
});
