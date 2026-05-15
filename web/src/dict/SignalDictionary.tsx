import React, { useState, useEffect, useMemo } from "react";
import type { SemanticSignal, AssignmentsStoreOptions } from "../types";
import type { SignalCatalogueAdapter } from "../types";
import { useAssignments } from "./useAssignments";

export type SignalDictionaryProps = {
  adapter: SignalCatalogueAdapter;
  store: AssignmentsStoreOptions;
  wallId: string;
  className?: string;
};

export function SignalDictionary({ adapter, store, wallId, className }: SignalDictionaryProps) {
  const [tab, setTab] = useState<"monitor" | "control">("monitor");
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<{ group: string; subgroup: string } | null>(null);
  const assigns = useAssignments(store);

  const catalogue = adapter.list();
  const channels = adapter.channels();

  const filtered = useMemo(
    () => catalogue.filter((p) => p.role === tab),
    [catalogue, tab]
  );

  const groups = useMemo(() => {
    const g: Record<string, Record<string, SemanticSignal[]>> = {};
    filtered.forEach((p) => {
      if (!g[p.group]) g[p.group] = {};
      if (!g[p.group][p.subgroup]) g[p.group][p.subgroup] = [];
      g[p.group][p.subgroup].push(p);
    });
    return g;
  }, [filtered]);

  useEffect(() => {
    if (selected && groups[selected.group]?.[selected.subgroup]) return;
    const firstGroup = Object.keys(groups)[0];
    if (firstGroup) {
      const firstSgrp = Object.keys(groups[firstGroup])[0];
      setSelected({ group: firstGroup, subgroup: firstSgrp });
    } else {
      setSelected(null);
    }
  }, [tab, groups]);

  const signals = selected && groups[selected.group]
    ? (groups[selected.group][selected.subgroup] || [])
    : [];

  const displayed = signals.filter(
    (s) => !query || s.name.toLowerCase().includes(query.toLowerCase()) || String(s.id).includes(query)
  );

  return (
    <div className={"sf-dict" + (className ? " " + className : "")}>
      <div className="sf-dict-rail">
        <div className="sf-dict-tabs">
          <button className={tab === "monitor" ? "active" : ""} onClick={() => setTab("monitor")}>Telemetry</button>
          <button className={tab === "control" ? "active" : ""} onClick={() => setTab("control")}>Telecommands</button>
        </div>
        <input className="sf-dict-search" placeholder="search signals…" value={query} onChange={(e) => setQuery(e.target.value)} />
        <div className="sf-dict-groups">
          {Object.entries(groups).map(([grp, sgrps]) => (
            <DictGroup key={grp} group={grp} sgrps={sgrps} selected={selected} onSelect={setSelected} query={query} />
          ))}
        </div>
      </div>
      <div className="sf-dict-main">
        {selected && (
          <div className="sf-dict-breadcrumb">
            <span>{selected.group}</span>
            <span className="sep">›</span>
            <span>{selected.subgroup}</span>
            <span className="sf-count">{signals.length} signals</span>
          </div>
        )}
        {displayed.map((signal) => (
          <SignalRow
            key={signal.id}
            signal={signal}
            channels={adapter.channelsForSignal(signal)}
            adapter={adapter}
            wallId={wallId}
            assigns={assigns}
            tab={tab}
          />
        ))}
        {signals.length === 0 && (
          <div className="sf-dict-empty">No signals here.</div>
        )}
      </div>
    </div>
  );
}

function DictGroup({
  group,
  sgrps,
  selected,
  onSelect,
  query,
}: {
  group: string;
  sgrps: Record<string, SemanticSignal[]>;
  selected: { group: string; subgroup: string } | null;
  onSelect: (s: { group: string; subgroup: string }) => void;
  query: string;
}) {
  const [open, setOpen] = useState(true);
  let visibleSgrps = sgrps;
  if (query) {
    const f: typeof sgrps = {};
    Object.entries(sgrps).forEach(([k, sigs]) => {
      const m = sigs.filter((s) => s.name.toLowerCase().includes(query.toLowerCase()) || String(s.id).includes(query));
      if (m.length) f[k] = m;
    });
    if (!Object.keys(f).length) return null;
    visibleSgrps = f;
  }
  const total = Object.values(visibleSgrps).reduce((a, b) => a + b.length, 0);
  return (
    <div className="sf-dict-group">
      <div className="sf-dict-group-head" onClick={() => setOpen((o) => !o)}>
        <span>{open ? "▾" : "▸"}</span>
        <span>{group}</span>
        <span className="sf-count">{total}</span>
      </div>
      {open && Object.entries(visibleSgrps).map(([sgrp, items]) => {
        const sel = selected?.group === group && selected?.subgroup === sgrp;
        return (
          <div key={sgrp} className={"sf-dict-sgrp" + (sel ? " selected" : "")}
               onClick={() => onSelect({ group, subgroup: sgrp })}>
            <span>{sgrp}</span>
            <span className="sf-count">{items.length}</span>
          </div>
        );
      })}
    </div>
  );
}

function SignalRow({
  signal,
  channels,
  adapter,
  wallId,
  assigns,
  tab,
}: {
  signal: SemanticSignal;
  channels: ReturnType<SignalCatalogueAdapter["channels"]>;
  adapter: SignalCatalogueAdapter;
  wallId: string;
  assigns: ReturnType<typeof useAssignments>;
  tab: string;
}) {
  const allAssigned = channels.every((c) => assigns.hasAssignment(wallId, signal.id, c.device_id, c.instance));
  function toggle() {
    if (allAssigned) {
      channels.forEach((c) => assigns.remove(wallId, signal.id, c.device_id, c.instance));
    } else {
      channels.forEach((c) => assigns.add(wallId, signal.id, c.device_id, c.instance));
    }
  }

  return (
    <div className="sf-signal-row">
      <div className="sf-signal-head">
        <span className="sf-signal-name">{signal.name}</span>
        <span className="sf-signal-meta">#{signal.id}{signal.unit ? ` · ${signal.unit}` : ""} · {signal.kind}</span>
      </div>
      {channels.length > 0 && tab === "monitor" && (
        <label className="sf-signal-assign">
          <input type="checkbox" checked={allAssigned} onChange={toggle} />
          <span className="sf-assign-swatch" style={{ background: adapter.colorForRole(signal.role) }}></span>
          Show in wall · {channels.length} ch
        </label>
      )}
    </div>
  );
}
