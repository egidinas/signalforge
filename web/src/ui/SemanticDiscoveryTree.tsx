import React from "react";
// Note: We avoid heavy dependencies like react-arborist in the core shared package 
// if possible, or we provide a lightweight alternative. 
// For this demonstrator, we'll implement a lightweight semantic tree.

export interface SemanticDiscoveryNode {
  id: string;
  label: string;
  kind: 'node' | 'device' | 'subsystem' | 'format' | 'target' | string;
  target_id?: string;
  family?: string;
  format?: string;
  owner_label?: string;
  use_label?: string;
  signal_count?: number;
  message_count?: number;
  badges?: string[];
  children?: SemanticDiscoveryNode[];
  sparkline_data?: number[];
}

export interface SemanticDiscoveryTreeProps {
  nodes: SemanticDiscoveryNode[];
  selectedID?: string;
  onSelect?: (node: SemanticDiscoveryNode) => void;
  renderExtra?: (node: SemanticDiscoveryNode) => React.ReactNode;
}

const KIND_LABELS: Record<string, string> = {
  node: 'ND',
  device: 'DV',
  subsystem: 'SS',
  format: 'FM',
  target: 'TG',
};

export function SemanticDiscoveryTree({
  nodes,
  selectedID,
  onSelect,
  renderExtra,
}: SemanticDiscoveryTreeProps) {
  return (
    <div className="sf-semantic-tree">
      {nodes.map(node => (
        <SemanticDiscoveryRow 
          key={node.id} 
          node={node} 
          level={0} 
          selectedID={selectedID} 
          onSelect={onSelect}
          renderExtra={renderExtra}
        />
      ))}
    </div>
  );
}

function SemanticDiscoveryRow({ node, level, selectedID, onSelect, renderExtra }: { 
  node: SemanticDiscoveryNode; 
  level: number; 
  selectedID?: string; 
  onSelect?: (n: SemanticDiscoveryNode) => void;
  renderExtra?: (n: SemanticDiscoveryNode) => React.ReactNode;
}) {
  const [open, setOpen] = React.useState(level < 1);
  const isSelected = node.id === selectedID;
  const hasChildren = node.children && node.children.length > 0;

  return (
    <div className="sf-semantic-row-container">
      <div 
        className={`sf-semantic-row sf-semantic-row--${node.kind} ${isSelected ? 'sf-selected' : ''}`}
        style={{ paddingLeft: level * 16 + 8 }}
        onClick={() => onSelect?.(node)}
      >
        <button 
          className="sf-semantic-twist" 
          onClick={(e) => { e.stopPropagation(); setOpen(!open); }}
          disabled={!hasChildren}
        >
          {hasChildren ? (open ? '▼' : '▶') : ''}
        </button>
        <span className="sf-semantic-kind">{KIND_LABELS[node.kind] ?? node.kind.slice(0, 2).toUpperCase()}</span>
        <span className="sf-semantic-label">{node.label}</span>
        {node.family && <span className="sf-semantic-pill">{node.family}</span>}
        {node.badges?.map(b => <span key={b} className="sf-semantic-badge">{b}</span>)}
        {renderExtra?.(node)}
      </div>
      {hasChildren && open && (
        <div className="sf-semantic-children">
          {node.children!.map(child => (
            <SemanticDiscoveryRow 
              key={child.id} 
              node={child} 
              level={level + 1} 
              selectedID={selectedID} 
              onSelect={onSelect}
              renderExtra={renderExtra}
            />
          ))}
        </div>
      )}
    </div>
  );
}
