import {
  channelSemanticTarget,
  normalizeSemanticOverlayBundle,
  overlayEntryForTarget,
  removeSemanticOverlayEntry,
  semanticOverlayTargetKey,
  upsertSemanticOverlayEntry,
} from "../src/catalogue/semanticOverlay";

describe("semanticOverlay", () => {
  test("builds stable keys for channel targets", () => {
    expect(semanticOverlayTargetKey(channelSemanticTarget({ device_id: "tec-76", instance: 1, role: "temp" }))).toBe("channel:tec-76:1:_:temp");
  });

  test("normalizes and deduplicates entries", () => {
    const bundle = normalizeSemanticOverlayBundle({
      namespace: "mecomgw",
      entries: [
        { target: channelSemanticTarget({ device_id: "tec-76", instance: 1, role: "temp" }), alias: " HR1 control " },
        { target: channelSemanticTarget({ device_id: "tec-76", instance: 1, role: "temp" }), alias: "Origin control" },
      ],
    }, "fallback");

    expect(bundle.namespace).toBe("mecomgw");
    expect(bundle.entries).toHaveLength(1);
    expect(bundle.entries[0].alias).toBe("Origin control");
  });

  test("upserts and removes entries by target", () => {
    const target = channelSemanticTarget({ device_id: "tec-76", instance: 2, role: "supply" });
    const withEntry = upsertSemanticOverlayEntry(null, { target, alias: "28 V supply", author: "operator" }, "mecomgw");

    expect(overlayEntryForTarget(withEntry, target)?.alias).toBe("28 V supply");
    expect(overlayEntryForTarget(removeSemanticOverlayEntry(withEntry, target), target)).toBeNull();
  });
});
