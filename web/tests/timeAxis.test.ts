import { clampRange, zoomRangeAroundAnchor } from "../src/render/timeAxis";

describe("timeAxis", () => {
  it("does not expand a clamped range beyond the available full span", () => {
    expect(clampRange({ start: 25, end: 50 }, { start: 0, end: 100 }, 1_000)).toEqual({ start: 0, end: 100 });
  });

  it("keeps the current time anchored when zooming", () => {
    expect(zoomRangeAroundAnchor({ start: 0, end: 100 }, { start: 0, end: 100 }, 0.5, 1, 100)).toEqual({ start: 50, end: 100 });
    expect(zoomRangeAroundAnchor({ start: 50, end: 100 }, { start: 0, end: 200 }, 2, 1, 100)).toEqual({ start: 0, end: 100 });
  });

  it("falls back to center zoom when there is no current-time anchor", () => {
    expect(zoomRangeAroundAnchor({ start: 0, end: 100 }, { start: 0, end: 100 }, 0.5, 1)).toEqual({ start: 25, end: 75 });
  });
});
