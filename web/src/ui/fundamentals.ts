export const SF_UI_CLASSNAMES = {
  semanticValuePopup: {
    root: "sf-semantic-value-popup",
    panel: "sf-semantic-value-popup__panel",
    row: "sf-semantic-value-popup__row",
    label: "sf-semantic-value-popup__label",
    value: "sf-semantic-value-popup__value",
  },
  writeLifecycleTrace: {
    root: "sf-write-lifecycle-trace",
    head: "sf-write-lifecycle-trace__head",
    title: "sf-write-lifecycle-trace__title",
    meta: "sf-write-lifecycle-trace__meta",
    values: "sf-write-lifecycle-trace__values",
    valueRow: "sf-write-lifecycle-trace__value-row",
    steps: "sf-write-lifecycle-trace__steps",
    step: "sf-write-lifecycle-trace__step",
  },
  densePanel: {
    root: "sf-dense-panel",
    rail: "sf-dense-panel__rail",
    body: "sf-dense-panel__body",
  },
  tileControls: {
    root: "sf-tile-controls",
    button: "sf-tile-controls__button",
    toggle: "sf-tile-controls__toggle",
    field: "sf-tile-controls__field",
  },
} as const;

export const SF_UI_TOKENS = {
  spacing: {
    dense: 8,
    compact: 6,
    related: 10,
  },
  radius: {
    panel: 8,
    control: 6,
    popup: 8,
  },
  typography: {
    denseFontSize: 12,
    valueFontSize: 13,
    labelFontSize: 11,
  },
  surfaces: {
    panelBackground: "var(--sf-surface-panel, #ffffff)",
    popupBackground: "var(--sf-surface-popup, #ffffff)",
    borderColor: "var(--sf-border-subtle, rgba(15, 23, 42, 0.14))",
    shadow: "var(--sf-shadow-popup, 0 16px 34px rgba(15, 23, 42, 0.16))",
  },
  densePanelGap: 8,
  densePanelRadius: 8,
  semanticPopupMinWidth: 220,
  semanticPopupMaxWidth: "min(420px, 68vw)",
  traceStepMinWidth: 96,
  tileControlHeight: 26,
} as const;
