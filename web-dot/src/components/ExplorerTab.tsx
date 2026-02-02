import React, { useState, useEffect, useCallback, useRef } from 'react';
import { guiClient } from '../adapters/GuiClient.js';
import type { NodeVM } from '../adapters/GuiClient.js';

type GraphNode = { id: string; title: string; isCenter: boolean; x: number; y: number; roles: string[] };
type GraphLink = { source: string; target: string; type: string };

type NodeShape = 'circle' | 'roundedRect' | 'hexagon' | 'diamond' | 'document' | 'shield';

type ExplorerTabProps = {
  contextNodeId: string | null;
  contextNamespaceId: string | null;
};

function getNodeShape(roles: string[]): NodeShape {
  const r = roles.join(' ').toLowerCase();
  if (/decision/i.test(r)) return 'diamond';
  if (/evidence/i.test(r) || /material/i.test(r)) return 'document';
  if (/assuranceobligation/i.test(r) || /assurance_obligation/i.test(r)) return 'shield';
  if (/fin\.policy/i.test(r) || (/\bpolicy\b/.test(r) && !/policy_set/i.test(r))) return 'shield';
  if (/transformationtheme/i.test(r)) return 'shield';
  if (/\bgoal\b/.test(r) || /objective/i.test(r) || /strategicobjective/i.test(r) || /strategic_objective/i.test(r) || /fin\.objective/i.test(r)) return 'roundedRect';
  if (/\bjob\b/.test(r)) return 'hexagon';
  if (/product\.domain/i.test(r) || /ledgerroot/i.test(r) || /ledgeraccount/i.test(r) || /subaccount/i.test(r) || /reservepool/i.test(r) || /reserveasset/i.test(r) || /treasuryaccount/i.test(r) || /workitem/i.test(r)) return 'circle';
  return 'circle';
}

function getNodeColor(shape: NodeShape, isCenter: boolean): { fill: string; stroke: string } {
  if (isCenter) return { fill: '#1565c0', stroke: '#0d47a1' };
  const palette: Record<NodeShape, { fill: string; stroke: string }> = {
    circle: { fill: '#e3f2fd', stroke: '#1976d2' },
    roundedRect: { fill: '#e8f5e9', stroke: '#388e3c' },
    hexagon: { fill: '#fff3e0', stroke: '#f57c00' },
    diamond: { fill: '#fce4ec', stroke: '#c2185b' },
    document: { fill: '#f3e5f5', stroke: '#7b1fa2' },
    shield: { fill: '#e0f2f1', stroke: '#00796b' },
  };
  return palette[shape];
}

function getLinkStyle(type: string): { stroke: string; strokeDasharray: string; markerId: string } {
  const t = type.toUpperCase();
  if (t === 'SUPPORTS' || t === 'SATISFIES' || t === 'ADVANCES') return { stroke: '#2e7d32', strokeDasharray: 'none', markerId: 'arrow-green' };
  if (t === 'DEPENDS_ON' || t === 'AFFECTS') return { stroke: '#1565c0', strokeDasharray: '6 4', markerId: 'arrow-blue' };
  if (t === 'DECIDED_BY' || t === 'EVIDENCED_BY') return { stroke: '#6a1b9a', strokeDasharray: '3 3', markerId: 'arrow-purple' };
  if (t === 'CONTAINS') return { stroke: '#424242', strokeDasharray: 'none', markerId: 'arrow' };
  return { stroke: '#757575', strokeDasharray: '4 2', markerId: 'arrow' };
}

function buildGraphFromNodeVM(centerId: string, vm: NodeVM): { nodes: GraphNode[]; links: GraphLink[] } {
  const nodes: GraphNode[] = [];
  const links: GraphLink[] = [];
  const seen = new Set<string>();

  nodes.push({
    id: vm.node.node_id,
    title: vm.node.title,
    isCenter: true,
    x: 0,
    y: 0,
    roles: vm.node.roles ?? [],
  });
  seen.add(vm.node.node_id);

  const addNeighbour = (rel: { node_id: string; title: string; roles?: string[] }, linkType: string) => {
    if (!rel?.node_id) return;
    if (!seen.has(rel.node_id)) {
      seen.add(rel.node_id);
      nodes.push({ id: rel.node_id, title: rel.title, isCenter: false, x: 0, y: 0, roles: rel.roles ?? [] });
    }
    links.push({ source: centerId, target: rel.node_id, type: linkType });
  };

  for (const c of vm.relationships.children) addNeighbour(c, 'CONTAINS');
  for (const a of vm.relationships.alignment) addNeighbour(a, a.type || 'alignment');
  for (const c of vm.relationships.coherence) addNeighbour(c, c.type || 'coherence');
  for (const d of vm.relationships.decisions_and_evidence) addNeighbour(d, d.type || 'decision');
  for (const o of vm.relationships.other ?? []) addNeighbour(o, o.type || 'other');

  return { nodes, links };
}

function NodeShapePath({
  shape,
  size,
  fill,
  stroke,
}: {
  shape: NodeShape;
  size: number;
  fill: string;
  stroke: string;
}) {
  const s = size;
  const w = s * 1.4;
  const h = s * 1.1;
  if (shape === 'circle') {
    return <circle r={s} fill={fill} stroke={stroke} strokeWidth={2} />;
  }
  if (shape === 'roundedRect') {
    return <rect x={-w / 2} y={-h / 2} width={w} height={h} rx={s * 0.25} ry={s * 0.25} fill={fill} stroke={stroke} strokeWidth={2} />;
  }
  if (shape === 'hexagon') {
    const a = (Math.PI / 3) * 0;
    const points = Array.from({ length: 6 }, (_, i) => {
      const angle = a + (Math.PI / 3) * i - Math.PI / 2;
      return `${s * 1.1 * Math.cos(angle)},${s * 1.1 * Math.sin(angle)}`;
    }).join(' ');
    return <polygon points={points} fill={fill} stroke={stroke} strokeWidth={2} />;
  }
  if (shape === 'diamond') {
    const points = `0,${-s} ${s},0 0,${s} ${-s},0`;
    return <polygon points={points} fill={fill} stroke={stroke} strokeWidth={2} />;
  }
  if (shape === 'document') {
    const r = s * 0.3;
    const path = `M ${-w / 2 + r},${-h / 2} L ${w / 2 - r},${-h / 2} Q ${w / 2},${-h / 2} ${w / 2},${-h / 2 + r} L ${w / 2},${h / 2 - r} Q ${w / 2},${h / 2} ${w / 2 - r},${h / 2} L ${-w / 2 + r},${h / 2} Q ${-w / 2},${h / 2} ${-w / 2},${h / 2 - r} L ${-w / 2},${-h / 2 + r} Q ${-w / 2},${-h / 2} ${-w / 2 + r},${-h / 2} Z`;
    return <path d={path} fill={fill} stroke={stroke} strokeWidth={2} />;
  }
  if (shape === 'shield') {
    const path = `M 0,${-s} L ${s * 1.1},${-s * 0.3} L ${s * 1.1},${s * 0.4} Q ${s * 1.1},${s} 0,${s} Q ${-s * 1.1},${s} ${-s * 1.1},${s * 0.4} L ${-s * 1.1},${-s * 0.3} Z`;
    return <path d={path} fill={fill} stroke={stroke} strokeWidth={2} />;
  }
  return <circle r={s} fill={fill} stroke={stroke} strokeWidth={2} />;
}

function layoutGraph(nodes: GraphNode[], links: GraphLink[]): GraphNode[] {
  const center = nodes.find((n) => n.isCenter);
  if (!center) return nodes;
  const centerId = center.id;
  const directIds = new Set<string>();
  for (const l of links) {
    if (l.source === centerId || l.target === centerId) {
      directIds.add(l.source === centerId ? l.target : l.source);
    }
  }
  const direct = nodes.filter((n) => !n.isCenter && directIds.has(n.id));
  const rest = nodes.filter((n) => !n.isCenter && !directIds.has(n.id));
  const R1 = 160;
  const R2 = 280;
  direct.forEach((n, i) => {
    const angle = (2 * Math.PI * i) / Math.max(direct.length, 1);
    n.x = center.x + R1 * Math.cos(angle);
    n.y = center.y + R1 * Math.sin(angle);
  });
  rest.forEach((n, i) => {
    const angle = (2 * Math.PI * i) / Math.max(rest.length, 1);
    n.x = center.x + R2 * Math.cos(angle);
    n.y = center.y + R2 * Math.sin(angle);
  });
  return nodes;
}

export default function ExplorerTab({ contextNodeId, contextNamespaceId }: ExplorerTabProps) {
  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [links, setLinks] = useState<GraphLink[]>([]);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [transform, setTransform] = useState({ x: 0, y: 0, k: 1 });
  const [dragging, setDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const svgRef = useRef<SVGSVGElement>(null);
  const linksRef = useRef<GraphLink[]>([]);
  linksRef.current = links;

  const loadNeighbourhood = useCallback(
    async (nodeId: string, replace: boolean) => {
      if (!contextNamespaceId || !contextNodeId) return;
      setLoading(true);
      setError(null);
      try {
        const vm = await guiClient.node(nodeId, contextNamespaceId);
        const { nodes: newNodes, links: newLinks } = buildGraphFromNodeVM(nodeId, vm);
        if (replace) {
          const withCenter = newNodes.map((n) => ({
            ...n,
            isCenter: n.id === contextNodeId,
          }));
          const laidOut = layoutGraph(withCenter, newLinks);
          setNodes(laidOut);
          setLinks(newLinks);
          linksRef.current = newLinks;
        } else {
          const linkKey = (l: GraphLink) => `${l.source}-${l.target}-${l.type}`;
          const currentLinks = linksRef.current;
          const nextLinks = [...currentLinks];
          const linkSeen = new Set(currentLinks.map(linkKey));
          for (const l of newLinks) {
            if (!linkSeen.has(linkKey(l))) {
              linkSeen.add(linkKey(l));
              nextLinks.push(l);
            }
          }
          setNodes((prev) => {
            const byId = new Map(prev.map((n) => [n.id, { ...n }]));
            for (const n of newNodes) {
              const isCenter = n.id === contextNodeId;
              const existing = byId.get(n.id);
              const roles = (n.roles?.length ? n.roles : existing?.roles ?? []) as string[];
              if (!byId.has(n.id)) byId.set(n.id, { ...n, isCenter, roles });
              else byId.set(n.id, { ...existing!, ...n, isCenter, roles });
            }
            const merged = Array.from(byId.values());
            return layoutGraph(merged, nextLinks);
          });
          setLinks(nextLinks);
          linksRef.current = nextLinks;
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load neighbourhood');
      } finally {
        setLoading(false);
      }
    },
    [contextNamespaceId, contextNodeId]
  );

  useEffect(() => {
    if (!contextNodeId || !contextNamespaceId) {
      setNodes([]);
      setLinks([]);
      setExpandedIds(new Set());
      linksRef.current = [];
      return;
    }
    setNodes([]);
    setLinks([]);
    setExpandedIds(new Set([contextNodeId]));
    linksRef.current = [];
    loadNeighbourhood(contextNodeId, true);
  }, [contextNodeId, contextNamespaceId, loadNeighbourhood]);

  const handleExpandNode = useCallback(
    (nodeId: string) => {
      if (expandedIds.has(nodeId)) return;
      setExpandedIds((prev) => new Set(prev).add(nodeId));
      loadNeighbourhood(nodeId, false);
    },
    [expandedIds, loadNeighbourhood]
  );

  if (!contextNamespaceId) {
    return (
      <div style={{ padding: '2rem', color: '#666', textAlign: 'center' }}>
        Select a node in the left nav to view its neighbourhood in the Explorer.
      </div>
    );
  }

  if (!contextNodeId) {
    return (
      <div style={{ padding: '2rem', color: '#666', textAlign: 'center' }}>
        Select a node in the left nav to view its neighbourhood in the Explorer.
      </div>
    );
  }

  if (loading && nodes.length === 0) {
    return <div className="loading">Loading neighbourhood…</div>;
  }

  if (error && nodes.length === 0) {
    return <div className="error">Error: {error}</div>;
  }

  const width = 800;
  const height = 500;
  const centerX = width / 2;
  const centerY = height / 2;

  const onWheel = (e: React.WheelEvent) => {
    e.preventDefault();
    const scale = e.deltaY > 0 ? 0.9 : 1.1;
    setTransform((t) => ({ ...t, k: Math.min(3, Math.max(0.2, t.k * scale)) }));
  };

  const onMouseDown = (e: React.MouseEvent) => {
    if (e.button === 0) {
      setDragging(true);
      setDragStart({ x: e.clientX - transform.x, y: e.clientY - transform.y });
    }
  };

  const onMouseMove = (e: React.MouseEvent) => {
    if (dragging) {
      setTransform((t) => ({ ...t, x: e.clientX - dragStart.x, y: e.clientY - dragStart.y }));
    }
  };

  const onMouseUp = () => setDragging(false);
  const onMouseLeave = () => setDragging(false);

  const legendShapes: { shape: NodeShape; label: string }[] = [
    { shape: 'circle', label: 'Domain / Product' },
    { shape: 'roundedRect', label: 'Goal / Objective' },
    { shape: 'hexagon', label: 'Job' },
    { shape: 'diamond', label: 'Decision' },
    { shape: 'document', label: 'Evidence / Material' },
    { shape: 'shield', label: 'Policy / Obligation' },
  ];
  const legendEdges: { type: string; label: string }[] = [
    { type: 'SUPPORTS', label: 'SUPPORTS / SATISFIES' },
    { type: 'DEPENDS_ON', label: 'DEPENDS_ON / AFFECTS' },
    { type: 'EVIDENCED_BY', label: 'DECIDED_BY / EVIDENCED_BY' },
    { type: 'CONTAINS', label: 'CONTAINS' },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', padding: '1rem' }}>
      <p style={{ fontSize: '0.875rem', color: '#666', marginBottom: '0.5rem' }}>
        1-hop neighbourhood of the selected node. Click a node to expand and add its connections (read-only).
      </p>
      <div
        style={{
          flex: 1,
          minHeight: 400,
          border: '1px solid #e0e0e0',
          borderRadius: '8px',
          overflow: 'hidden',
          background: '#fafafa',
        }}
        onWheel={onWheel}
        onMouseDown={onMouseDown}
        onMouseMove={onMouseMove}
        onMouseUp={onMouseUp}
        onMouseLeave={onMouseLeave}
      >
        <svg
          ref={svgRef}
          width="100%"
          height="100%"
          viewBox={`0 0 ${width} ${height}`}
          style={{ cursor: dragging ? 'grabbing' : 'grab' }}
        >
          <defs>
            <marker id="arrow" markerWidth={8} markerHeight={8} refX={6} refY={3} orient="auto">
              <path d="M0,0 L6,3 L0,6 Z" fill="#616161" />
            </marker>
            <marker id="arrow-green" markerWidth={8} markerHeight={8} refX={6} refY={3} orient="auto">
              <path d="M0,0 L6,3 L0,6 Z" fill="#2e7d32" />
            </marker>
            <marker id="arrow-blue" markerWidth={8} markerHeight={8} refX={6} refY={3} orient="auto">
              <path d="M0,0 L6,3 L0,6 Z" fill="#1565c0" />
            </marker>
            <marker id="arrow-purple" markerWidth={8} markerHeight={8} refX={6} refY={3} orient="auto">
              <path d="M0,0 L6,3 L0,6 Z" fill="#6a1b9a" />
            </marker>
          </defs>
          <g transform={`translate(${centerX + transform.x},${centerY + transform.y}) scale(${transform.k})`}>
            {links.map((l, i) => {
              const src = nodes.find((n) => n.id === l.source);
              const tgt = nodes.find((n) => n.id === l.target);
              if (!src || !tgt) return null;
              const style = getLinkStyle(l.type);
              return (
                <line
                  key={`${l.source}-${l.target}-${l.type}-${i}`}
                  x1={src.x}
                  y1={src.y}
                  x2={tgt.x}
                  y2={tgt.y}
                  stroke={style.stroke}
                  strokeWidth={2}
                  strokeDasharray={style.strokeDasharray}
                  markerEnd={`url(#${style.markerId})`}
                />
              );
            })}
            {nodes.map((n) => {
              const shape = getNodeShape(n.roles);
              const colors = getNodeColor(shape, n.isCenter);
              const size = n.isCenter ? 20 : 14;
              return (
                <g
                  key={n.id}
                  transform={`translate(${n.x},${n.y})`}
                  style={{ cursor: 'pointer' }}
                  onMouseDown={(e) => e.stopPropagation()}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleExpandNode(n.id);
                  }}
                >
                  <NodeShapePath shape={shape} size={size} fill={colors.fill} stroke={colors.stroke} />
                  <text
                    textAnchor="middle"
                    dy={size + (n.isCenter ? 14 : 10)}
                    fontSize={n.isCenter ? 11 : 10}
                    fill="#333"
                    style={{ pointerEvents: 'none', userSelect: 'none' }}
                  >
                    {n.title.length > 24 ? n.title.slice(0, 21) + '…' : n.title}
                  </text>
                </g>
              );
            })}
          </g>
        </svg>
      </div>
      <div
        style={{
          marginTop: '0.75rem',
          padding: '0.75rem 1rem',
          background: '#fff',
          border: '1px solid #e0e0e0',
          borderRadius: '8px',
          fontSize: '0.75rem',
          display: 'flex',
          flexWrap: 'wrap',
          gap: '1rem',
          alignItems: 'center',
        }}
      >
        <span style={{ fontWeight: 600, color: '#555', marginRight: '0.25rem' }}>Shapes:</span>
        {legendShapes.map(({ shape, label }) => {
          const colors = getNodeColor(shape, false);
          return (
            <span key={shape} style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
              <svg width={20} height={20} viewBox="-12 -12 24 24" style={{ flexShrink: 0 }}>
                <NodeShapePath shape={shape} size={10} fill={colors.fill} stroke={colors.stroke} />
              </svg>
              <span style={{ color: '#333' }}>{label}</span>
            </span>
          );
        })}
        <span style={{ fontWeight: 600, color: '#555', marginLeft: '0.5rem', marginRight: '0.25rem' }}>Edges:</span>
        {legendEdges.map(({ type, label }) => {
          const style = getLinkStyle(type);
          return (
            <span key={type} style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
              <svg width={32} height={10} style={{ flexShrink: 0 }}>
                <line x1={2} y1={5} x2={28} y2={5} stroke={style.stroke} strokeWidth={2} strokeDasharray={style.strokeDasharray} />
                <path d="M28,5 L24,2 L24,8 Z" fill={style.stroke} />
              </svg>
              <span style={{ color: '#333' }}>{label}</span>
            </span>
          );
        })}
      </div>
    </div>
  );
}
