export type SignalProjectionSignalID = string | number;

export const CURRENT_SOURCE_PROJECTION_SCHEMA_VERSION = 1;

export type SignalProjectionSourceFamily =
  | "mecom_tec"
  | "tec"
  | "opcua"
  | "can_dbc"
  | "ni_shared_variable"
  | "niagara"
  | "derived"
  | "other_discovered"
  | string;

export type SignalProjectionPathSegment = {
  id: string;
  label: string;
  kind?: string;
  order?: number;
  default_collapsed?: boolean;
  description?: string;
  source_family?: SignalProjectionSourceFamily;
  metadata?: Record<string, unknown>;
};

export type SignalProjectionSignalRef = {
  signal_id?: SignalProjectionSignalID;
  sid?: string;
  trace_id?: string;
  source_id?: string;
  source_family?: SignalProjectionSourceFamily;
  target_id?: string;
  role?: string;
  unit?: string;
  metadata?: Record<string, unknown>;
};

export type SignalProjectionMappingKind = "primary" | "secondary" | "hidden" | "preview" | string;

export type SignalProjectionMapping = {
  id?: string;
  bundle?: string;
  kind: SignalProjectionMappingKind;
  path: SignalProjectionPathSegment[];
  signal_refs: SignalProjectionSignalRef[];
  group_key?: string;
  device_grouping?: "device" | "channel" | "source" | "none" | string;
  sort_key?: string;
  default_visible?: boolean;
  default_collapsed?: boolean;
  reason?: string;
  source_id?: string;
  source_family?: SignalProjectionSourceFamily;
  title?: string;
  description?: string;
  metadata?: Record<string, unknown>;
};

export type SignalProjectionBundle = {
  schema_version?: string | number;
  namespace?: string;
  source?: string;
  generated_at?: string;
  default_grouping?: "device" | "channel" | "source" | "path" | string;
  mappings: SignalProjectionMapping[];
  metadata?: Record<string, unknown>;
};

export type SignalProjectionPathSegmentInput =
  | string
  | (Partial<SignalProjectionPathSegment> & { meta?: Record<string, unknown> });

export type SignalProjectionMappingInput = Partial<Omit<SignalProjectionMapping, "path" | "signal_refs" | "kind">> & {
  kind?: SignalProjectionMappingKind;
  path?: SignalProjectionPathSegmentInput[];
  signal_refs?: Array<Partial<SignalProjectionSignalRef> & { meta?: Record<string, unknown> }>;
  signal_ids?: SignalProjectionSignalID[];
  ids?: SignalProjectionSignalID[];
  trace_ids?: string[];
  meta?: Record<string, unknown>;
};

export type SignalProjectionBundleInput = Partial<Omit<SignalProjectionBundle, "mappings">> & {
  mappings?: SignalProjectionMappingInput[];
  primary_mappings?: SignalProjectionMappingInput[];
  secondary_mappings?: SignalProjectionMappingInput[];
  hidden_mappings?: SignalProjectionMappingInput[];
  meta?: Record<string, unknown>;
};

export type SignalProjectionValidationIssue = {
  code: string;
  message: string;
  mapping_id?: string;
  ref_key?: string;
  path_key?: string;
};

export type SignalProjectionValidationResult = {
  valid: boolean;
  errors: SignalProjectionValidationIssue[];
  warnings: SignalProjectionValidationIssue[];
};

export type SignalProjectionValidationOptions = {
  availableSignalIds?: SignalProjectionSignalID[];
  requiredPrimarySignalIds?: SignalProjectionSignalID[];
  catalogue?: { signals?: Array<Record<string, unknown>> };
  sourceCatalogues?: SignalProjectionSourceCatalogueLike[];
  requiredVisibility?: string;
  requireSecondaryReason?: boolean;
  allowPrimaryDuplicates?: boolean;
  allowUnknownSourceRefs?: boolean;
};

export type SignalProjectionSourceCatalogueLike = {
  source_id?: string;
  source_family?: SignalProjectionSourceFamily;
  entries?: SignalProjectionSourceCatalogueEntryLike[];
};

export type SignalProjectionSourceCatalogueEntryLike = {
  trace_id?: string;
  group_key?: string;
  group_label?: string;
  instance_key?: string;
  sort_key?: string;
  counterpart_group?: string;
  counterpart_trace_ids?: string[];
};

export type SignalProjectionTreeNode = {
  id: string;
  label: string;
  kind?: string;
  path: SignalProjectionPathSegment[];
  order?: number;
  default_collapsed?: boolean;
  mappings: SignalProjectionMapping[];
  children: SignalProjectionTreeNode[];
};

type SignalProjectionNodeLookup = {
  get(id: string): SignalProjectionTreeNode | undefined;
  set(id: string, node: SignalProjectionTreeNode): unknown;
};

export function normalizeSignalProjectionBundle(input: SignalProjectionBundleInput): SignalProjectionBundle {
  const mappings = [
    ...normalizeMappingGroup(input.mappings, undefined),
    ...normalizeMappingGroup(input.primary_mappings, "primary"),
    ...normalizeMappingGroup(input.secondary_mappings, "secondary"),
    ...normalizeMappingGroup(input.hidden_mappings, "hidden"),
  ];
  return {
    schema_version: input.schema_version ?? CURRENT_SOURCE_PROJECTION_SCHEMA_VERSION,
    namespace: optionalText(input.namespace),
    source: optionalText(input.source),
    generated_at: optionalText(input.generated_at),
    default_grouping: optionalText(input.default_grouping),
    mappings,
    metadata: metadataRecord(input),
  };
}

export function signalProjectionRefKey(ref: SignalProjectionSignalRef): string {
  if (ref.signal_id !== undefined && ref.signal_id !== null && String(ref.signal_id) !== "") return `id:${String(ref.signal_id)}`;
  if (ref.sid) return `sid:${ref.sid}`;
  if (ref.trace_id) return `trace:${ref.source_family || ""}:${ref.source_id || ""}:${ref.trace_id}`;
  if (ref.target_id) return `target:${ref.target_id}`;
  return "";
}

export function signalProjectionPathKey(path: SignalProjectionPathSegment[]): string {
  return path.map((segment) => segment.id).join("/");
}

export function validateSignalProjectionBundle(
  bundle: SignalProjectionBundle,
  options: SignalProjectionValidationOptions = {},
): SignalProjectionValidationResult {
  const errors: SignalProjectionValidationIssue[] = [];
  const warnings: SignalProjectionValidationIssue[] = [];
  if (Number(bundle.schema_version) !== CURRENT_SOURCE_PROJECTION_SCHEMA_VERSION || String(bundle.schema_version).trim() === "") {
    errors.push({
      code: "schema_version",
      message: `signal projection schema_version must be ${CURRENT_SOURCE_PROJECTION_SCHEMA_VERSION}`,
    });
  }
  const available = new Set<string>([
    ...keysFromIds(options.availableSignalIds || []),
    ...keysFromCatalogue(options.catalogue),
  ]);
  const availableTraceRefs = new Set<string>(keysFromSourceCatalogues(options.sourceCatalogues));
  const requiredPrimary = new Set<string>([
    ...keysFromIds(options.requiredPrimarySignalIds || []),
    ...requiredKeysFromCatalogue(options.catalogue, options.requiredVisibility || "operator"),
  ]);
  const primarySeen = new Map<string, SignalProjectionMapping>();

  for (const mapping of bundle.mappings) {
    const mappingID = mapping.id || mapping.bundle;
    const pathKey = signalProjectionPathKey(mapping.path);
    if (mapping.path.length === 0) {
      errors.push({ code: "missing_path", message: "projection mapping must have a non-empty path", mapping_id: mappingID });
    }
    if (mapping.signal_refs.length === 0) {
      errors.push({ code: "missing_signal_refs", message: "projection mapping must reference at least one signal", mapping_id: mappingID, path_key: pathKey });
    }
    if (mapping.kind === "secondary" && options.requireSecondaryReason !== false && String(mapping.reason || "").trim().length < 12) {
      errors.push({ code: "missing_secondary_reason", message: "secondary mappings require a review reason", mapping_id: mappingID, path_key: pathKey });
    }
    for (const ref of mapping.signal_refs) {
      const key = signalProjectionRefKey(ref);
      if (!key) {
        errors.push({ code: "empty_signal_ref", message: "projection signal reference must include signal_id, sid, trace_id, or target_id", mapping_id: mappingID, path_key: pathKey });
        continue;
      }
      if (available.size > 0 && key.startsWith("id:") && !available.has(key)) {
        errors.push({ code: "unknown_signal", message: `projection maps unknown signal ${key}`, mapping_id: mappingID, ref_key: key, path_key: pathKey });
      }
      if (availableTraceRefs.size > 0 && key.startsWith("trace:") && !options.allowUnknownSourceRefs && !availableTraceRefs.has(key)) {
        errors.push({ code: "unknown_source_ref", message: `projection maps unknown source trace ${key}`, mapping_id: mappingID, ref_key: key, path_key: pathKey });
      }
      if (mapping.kind === "primary") {
        requiredPrimary.delete(key);
        if (!options.allowPrimaryDuplicates && primarySeen.has(key)) {
          errors.push({ code: "duplicate_primary", message: `primary projection duplicates ${key}`, mapping_id: mappingID, ref_key: key, path_key: pathKey });
        }
        primarySeen.set(key, mapping);
      }
    }
  }

  for (const key of [...requiredPrimary].sort()) {
    errors.push({ code: "missing_primary", message: `required signal ${key} has no primary projection`, ref_key: key });
  }

  return { valid: errors.length === 0, errors, warnings };
}

export function buildSignalProjectionTree(bundle: SignalProjectionBundle): SignalProjectionTreeNode[] {
  const roots = new Map<string, SignalProjectionTreeNode>();
  for (const mapping of bundle.mappings) {
    let level: SignalProjectionNodeLookup = roots;
    let path: SignalProjectionPathSegment[] = [];
    let node: SignalProjectionTreeNode | undefined;
    for (const segment of mapping.path) {
      path = [...path, segment];
      node = level.get(segment.id);
      if (!node) {
        node = {
          id: segment.id,
          label: segment.label,
          kind: segment.kind,
          path,
          order: segment.order,
          default_collapsed: segment.default_collapsed,
          mappings: [],
          children: [],
        };
        level.set(segment.id, node);
      }
      level = childrenMap(node);
    }
    if (node) node.mappings.push(mapping);
  }
  return sortTree([...roots.values()]);
}

function normalizeMappingGroup(mappings: SignalProjectionMappingInput[] | undefined, forcedKind: SignalProjectionMappingKind | undefined): SignalProjectionMapping[] {
  if (!Array.isArray(mappings)) return [];
  return mappings.map((mapping) => normalizeMapping(mapping, forcedKind)).filter((mapping): mapping is SignalProjectionMapping => mapping !== null);
}

function normalizeMapping(mapping: SignalProjectionMappingInput, forcedKind: SignalProjectionMappingKind | undefined): SignalProjectionMapping | null {
  if (!mapping || typeof mapping !== "object") return null;
  const path = Array.isArray(mapping.path) ? mapping.path.map(normalizePathSegment).filter((segment): segment is SignalProjectionPathSegment => segment !== null) : [];
  const signalRefs = normalizeSignalRefs(mapping);
  return {
    id: optionalText(mapping.id),
    bundle: optionalText(mapping.bundle),
    kind: forcedKind || mapping.kind || "primary",
    path,
    signal_refs: signalRefs,
    group_key: optionalText(mapping.group_key),
    device_grouping: optionalText(mapping.device_grouping),
    sort_key: optionalText(mapping.sort_key),
    default_visible: booleanOrUndefined(mapping.default_visible),
    default_collapsed: booleanOrUndefined(mapping.default_collapsed),
    reason: optionalText(mapping.reason),
    source_id: optionalText(mapping.source_id),
    source_family: optionalText(mapping.source_family),
    title: optionalText(mapping.title),
    description: optionalText(mapping.description),
    metadata: metadataRecord(mapping),
  };
}

function normalizeSignalRefs(mapping: SignalProjectionMappingInput): SignalProjectionSignalRef[] {
  const refs: SignalProjectionSignalRef[] = [];
  if (Array.isArray(mapping.signal_refs)) {
    for (const ref of mapping.signal_refs) {
      if (!ref || typeof ref !== "object") continue;
      refs.push({
        signal_id: ref.signal_id,
        sid: optionalText(ref.sid),
        trace_id: optionalText(ref.trace_id),
        source_id: optionalText(ref.source_id || mapping.source_id),
        source_family: optionalText(ref.source_family || mapping.source_family),
        target_id: optionalText(ref.target_id),
        role: optionalText(ref.role),
        unit: optionalText(ref.unit),
        metadata: metadataRecord(ref),
      });
    }
  }
  for (const id of [...(mapping.signal_ids || []), ...(mapping.ids || [])]) {
    refs.push({
      signal_id: id,
      source_id: optionalText(mapping.source_id),
      source_family: optionalText(mapping.source_family),
    });
  }
  for (const traceID of mapping.trace_ids || []) {
    refs.push({
      trace_id: traceID,
      source_id: optionalText(mapping.source_id),
      source_family: optionalText(mapping.source_family),
    });
  }
  return dedupeSignalRefs(refs);
}

function normalizePathSegment(segment: SignalProjectionPathSegmentInput): SignalProjectionPathSegment | null {
  if (typeof segment === "string") {
    const label = segment.trim();
    if (!label) return null;
    return { id: slug(label), label };
  }
  if (!segment || typeof segment !== "object") return null;
  const label = text(segment.label, text(segment.id, "group"));
  return {
    id: text(segment.id, slug(label)),
    label,
    kind: optionalText(segment.kind),
    order: numberOrUndefined(segment.order),
    default_collapsed: booleanOrUndefined(segment.default_collapsed),
    description: optionalText(segment.description),
    source_family: optionalText(segment.source_family),
    metadata: metadataRecord(segment),
  };
}

function metadataRecord(value: { metadata?: unknown; meta?: unknown }) {
  if (isRecord(value.metadata)) return value.metadata;
  if (isRecord(value.meta)) return value.meta;
  return undefined;
}

function dedupeSignalRefs(refs: SignalProjectionSignalRef[]) {
  const seen = new Set<string>();
  const out: SignalProjectionSignalRef[] = [];
  for (const ref of refs) {
    const key = signalProjectionRefKey(ref);
    if (!key || seen.has(key)) continue;
    seen.add(key);
    out.push(ref);
  }
  return out;
}

function keysFromIds(ids: SignalProjectionSignalID[]) {
  return ids.map((id) => `id:${String(id)}`);
}

function keysFromCatalogue(catalogue: SignalProjectionValidationOptions["catalogue"]) {
  if (!catalogue || !Array.isArray(catalogue.signals)) return [];
  return catalogue.signals.flatMap((signal) => {
    const keys: string[] = [];
    if (signal.id !== undefined && signal.id !== null) keys.push(`id:${String(signal.id)}`);
    if (typeof signal.sid === "string" && signal.sid.trim()) keys.push(`sid:${signal.sid}`);
    return keys;
  });
}

function keysFromSourceCatalogues(catalogues: SignalProjectionValidationOptions["sourceCatalogues"]) {
  if (!Array.isArray(catalogues)) return [];
  return catalogues.flatMap((catalogue) => {
    if (!catalogue || !Array.isArray(catalogue.entries)) return [];
    return catalogue.entries.flatMap((entry) => {
      const traceID = optionalText(entry?.trace_id);
      if (!traceID) return [];
      const family = optionalText(catalogue.source_family) || "";
      const sourceID = optionalText(catalogue.source_id) || "";
      return [
        `trace:${family}:${sourceID}:${traceID}`,
        `trace::${sourceID}:${traceID}`,
        `trace:${family}::${traceID}`,
        `trace:::${traceID}`,
      ];
    });
  });
}

function requiredKeysFromCatalogue(catalogue: SignalProjectionValidationOptions["catalogue"], requiredVisibility: string) {
  if (!catalogue || !Array.isArray(catalogue.signals)) return [];
  return catalogue.signals
    .filter((signal) => signal.visibility === requiredVisibility)
    .flatMap((signal) => signal.id !== undefined && signal.id !== null ? [`id:${String(signal.id)}`] : []);
}

function childrenMap(node: SignalProjectionTreeNode) {
  const map = new Map<string, SignalProjectionTreeNode>();
  for (const child of node.children) map.set(child.id, child);
  return {
    get: (id: string) => map.get(id),
    set: (id: string, child: SignalProjectionTreeNode) => {
      map.set(id, child);
      node.children = [...map.values()];
      return map;
    },
  };
}

function sortTree(nodes: SignalProjectionTreeNode[]): SignalProjectionTreeNode[] {
  return nodes
    .map((node) => ({ ...node, children: sortTree(node.children) }))
    .sort((a, b) => (a.order ?? Number.MAX_SAFE_INTEGER) - (b.order ?? Number.MAX_SAFE_INTEGER) || a.label.localeCompare(b.label) || a.id.localeCompare(b.id));
}

function slug(value: string) {
  return value.trim().toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "group";
}

function text(value: unknown, fallback = "") {
  if (value === undefined || value === null || value === "") return fallback;
  return String(value);
}

function optionalText(value: unknown) {
  if (value === undefined || value === null || value === "") return undefined;
  return String(value);
}

function numberOrUndefined(value: unknown) {
  if (value === undefined || value === null || value === "") return undefined;
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : undefined;
}

function booleanOrUndefined(value: unknown) {
  if (typeof value === "boolean") return value;
  return undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
