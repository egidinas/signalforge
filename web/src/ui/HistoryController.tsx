import React, { useState, useEffect } from "react";

export type HistoryRange = {
  id: string;
  label: string;
  ms: number;
};

export type HistoryAvailability = {
  earliest: string;
  latest: string;
};

export type HistoryControllerProps = {
  onRangeChange: (t0: string, t1: string) => void;
  isLive: boolean;
  onSetLive: (live: boolean) => void;
  availability?: HistoryAvailability[];
  fetchAvailability?: () => Promise<{ availability: HistoryAvailability[] }>;
  ranges?: HistoryRange[];
};

const DEFAULT_RANGES: HistoryRange[] = [
  { id: "15m", label: "15m", ms: 15 * 60_000 },
  { id: "1h",  label: "1h",  ms: 3600_000 },
  { id: "6h",  label: "6h",  ms: 6 * 3600_000 },
  { id: "24h", label: "24h", ms: 24 * 3600_000 },
  { id: "3d",  label: "3d",  ms: 3 * 24 * 3600_000 },
];

export function HistoryController({
  onRangeChange,
  isLive,
  onSetLive,
  availability: propAvailability,
  fetchAvailability,
  ranges = DEFAULT_RANGES,
}: HistoryControllerProps) {
  const [mode, setMode] = useState(isLive ? "live" : "history");
  const [range, setRange] = useState("1h");
  const [customT0, setCustomT0] = useState("");
  const [customT1, setCustomT1] = useState("");
  const [availability, setAvailability] = useState<HistoryAvailability[]>(propAvailability || []);

  useEffect(() => {
    if (mode === "history" && fetchAvailability && !propAvailability) {
      fetchAvailability().then((res) => {
        if (res && res.availability) setAvailability(res.availability);
      }).catch(() => {});
    }
  }, [mode, fetchAvailability, propAvailability]);

  useEffect(() => {
     setMode(isLive ? "live" : "history");
  }, [isLive]);

  function handleSetRange(id: string) {
    setRange(id);
    const r = ranges.find(x => x.id === id);
    if (r) {
      const t1 = new Date();
      const t0 = new Date(t1.getTime() - r.ms);
      onRangeChange(t0.toISOString(), t1.toISOString());
      onSetLive(false);
    }
  }

  function handleApplyCustom() {
    if (customT0 && customT1) {
      onRangeChange(new Date(customT0).toISOString(), new Date(customT1).toISOString());
      onSetLive(false);
    }
  }

  const overallEarliest = availability.length ? new Date(Math.min(...availability.map(a => new Date(a.earliest).getTime()))) : null;
  const overallLatest = availability.length ? new Date(Math.max(...availability.map(a => new Date(a.latest).getTime()))) : null;

  return (
    <div className="sf-history-ctrl">
      <div className="sf-mode-toggle">
        <button className={mode === "live" ? "active" : ""} onClick={() => { setMode("live"); onSetLive(true); }}>STREAMING</button>
        <button className={mode === "history" ? "active" : ""} onClick={() => { setMode("history"); onSetLive(false); }}>HISTORICAL</button>
      </div>

      {mode === "history" && (
        <div className="sf-history-options">
          <div className="sf-presets">
            {ranges.map(r => (
              <button key={r.id} className={range === r.id ? "active" : ""} onClick={() => handleSetRange(r.id)}>{r.label}</button>
            ))}
          </div>
          <div className="sf-custom">
            <input type="datetime-local" value={customT0} onChange={e => setCustomT0(e.target.value)} />
            <span>to</span>
            <input type="datetime-local" value={customT1} onChange={e => setCustomT1(e.target.value)} />
            <button className="sf-btn sm primary" onClick={handleApplyCustom}>Apply</button>
          </div>
          {overallEarliest && (
            <div className="sf-availability-info">
              <span className="dim">Available from</span>
              <span className="sf-availability-pill">{overallEarliest.toLocaleString()}</span>
              <span className="dim">to</span>
              <span className="sf-availability-pill">{overallLatest.toLocaleString()}</span>
            </div>
          )}
        </div>
      )}

      {mode === "live" && (
        <div className="sf-live-status">
          <span className="sf-live-dot"></span>
          <span className="sf-live-text">Real-time streaming enabled</span>
        </div>
      )}
    </div>
  );
}
