import packageJson from "../package.json";
import { SF_UI_CLASSNAMES, SF_UI_TOKENS } from "../src/ui/fundamentals";

describe("ui fundamentals", () => {
  it("exports reusable semantic UI class names and tokens", () => {
    expect(SF_UI_CLASSNAMES.semanticValuePopup.panel).toBe("sf-semantic-value-popup__panel");
    expect(SF_UI_CLASSNAMES.writeLifecycleTrace.values).toBe("sf-write-lifecycle-trace__values");
    expect(SF_UI_CLASSNAMES.tileControls.field).toBe("sf-tile-controls__field");
    expect(SF_UI_TOKENS.radius.panel).toBe(8);
    expect(SF_UI_TOKENS.spacing.dense).toBe(8);
  });

  it("ships the package CSS entry referenced by exports", () => {
    expect(packageJson.exports["./styles.css"].default).toBe("./src/styles.css");
    expect(packageJson.files).toContain("src/styles.css");
  });
});
