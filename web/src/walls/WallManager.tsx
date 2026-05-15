import React, { useState } from "react";
import type { WallsHandle } from "./useWalls";

export type WallManagerProps = {
  walls: WallsHandle;
  selectedWallId?: string;
  onSelect: (wallId: string) => void;
};

export function WallManager({ walls, selectedWallId, onSelect }: WallManagerProps) {
  const [newLabel, setNewLabel] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editLabel, setEditLabel] = useState("");

  function addWall() {
    const label = newLabel.trim();
    if (!label) return;
    const wall = walls.add(label);
    setNewLabel("");
    onSelect(wall.wall_id);
  }

  function startEdit(wallId: string, label: string) {
    setEditingId(wallId);
    setEditLabel(label);
  }

  function commitEdit() {
    if (editingId && editLabel.trim()) {
      walls.rename(editingId, editLabel.trim());
    }
    setEditingId(null);
    setEditLabel("");
  }

  return (
    <div className="sf-wall-manager">
      <div className="sf-wall-list">
        {walls.walls.map((w) => (
          <div key={w.wall_id}
               className={"sf-wall-item" + (w.wall_id === selectedWallId ? " selected" : "") + (w.preset ? " preset" : "")}
               onClick={() => onSelect(w.wall_id)}>
            {editingId === w.wall_id ? (
              <input
                autoFocus
                value={editLabel}
                onChange={(e) => setEditLabel(e.target.value)}
                onBlur={commitEdit}
                onKeyDown={(e) => { if (e.key === "Enter") commitEdit(); if (e.key === "Escape") setEditingId(null); }}
                onClick={(e) => e.stopPropagation()}
              />
            ) : (
              <span className="sf-wall-label">{w.label}</span>
            )}
            {!w.preset && editingId !== w.wall_id && (
              <span className="sf-wall-actions">
                <button onClick={(e) => { e.stopPropagation(); startEdit(w.wall_id, w.label); }}>✎</button>
                <button onClick={(e) => { e.stopPropagation(); walls.remove(w.wall_id); }}>✕</button>
              </span>
            )}
          </div>
        ))}
        {walls.walls.length === 0 && (
          <div className="sf-wall-empty">No walls yet. Create one below.</div>
        )}
      </div>
      <div className="sf-wall-add">
        <input
          placeholder="New wall label…"
          value={newLabel}
          onChange={(e) => setNewLabel(e.target.value)}
          onKeyDown={(e) => { if (e.key === "Enter") addWall(); }}
        />
        <button onClick={addWall} disabled={!newLabel.trim()}>+ Add wall</button>
      </div>
    </div>
  );
}
