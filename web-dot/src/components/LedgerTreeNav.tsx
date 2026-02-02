import React, { useState, useEffect, useCallback } from 'react';
import { guiClient, type ProductTreeVM } from '../adapters/GuiClient.js';
import type { Node } from '../types/domain.js';
import { getNodeType, getNodeTypeColor, getNodeTypeShort, type NodeTypeLabel } from '../utils/nodeType.js';

type RootEntry = { namespaceId: string; rootNodeId: string; label: string };

type TreeItem = {
  node: Node;
  children: Node[];
  level: number;
  expanded: boolean;
  namespaceId: string;
  expandable: boolean;
  rootNodeId?: string;
  nodeType?: NodeTypeLabel | null;
  signals?: { alignment: boolean; coherence: boolean; decision: boolean; materials: boolean };
};

type LedgerTreeNavProps = {
  selectedNodeId: string | null;
  contextNamespaceId: string | null;
  onSelectNode: (nodeId: string, namespaceId: string) => void;
  pendingNodeId?: string | null;
  onClearPendingNavigation?: () => void;
};

function getSignals(
  nodeId: string,
  treeNode: ProductTreeVM['tree'] | null
): { alignment: boolean; coherence: boolean; decision: boolean; materials: boolean } | undefined {
  if (!treeNode) return undefined;
  if (treeNode.node_id === nodeId) return treeNode.signals;
  for (const child of treeNode.children || []) {
    const found = getSignals(nodeId, child);
    if (found) return found;
  }
  return undefined;
}

export default function LedgerTreeNav({
  selectedNodeId,
  contextNamespaceId,
  onSelectNode,
  pendingNodeId,
  onClearPendingNavigation,
}: LedgerTreeNavProps) {
  const [roots, setRoots] = useState<RootEntry[]>([]);
  const [rootTrees, setRootTrees] = useState<Map<string, { tree: ProductTreeVM['tree']; namespaceId: string }>>(new Map());
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [loadingRoots, setLoadingRoots] = useState(true);
  const [loadingRootId, setLoadingRootId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoadingRoots(true);
    setError(null);
    guiClient
      .namespaces()
      .then((data) => {
        const list = (data.namespaces || [])
          .map((n) => ({
            namespaceId: n.id,
            rootNodeId: n.root_node_id || '',
            label: n.label || n.id,
          }))
          .sort((a, b) => a.namespaceId.localeCompare(b.namespaceId));
        setRoots(list);
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load namespaces'))
      .finally(() => setLoadingRoots(false));
  }, []);

  const loadTreeForRoot = useCallback(async (rootNodeId: string, namespaceId: string) => {
    if (rootTrees.has(rootNodeId)) return;
    setLoadingRootId(rootNodeId);
    try {
      const vm = await guiClient.productsTree(rootNodeId, 10, namespaceId);
      if (vm.tree) {
        setRootTrees((prev) => {
          const next = new Map(prev);
          next.set(rootNodeId, { tree: vm.tree!, namespaceId });
          return next;
        });
        setExpandedNodes((prev) => new Set(prev).add(rootNodeId));
      }
    } catch (err) {
      console.error('Error loading tree for root:', err);
    } finally {
      setLoadingRootId(null);
    }
  }, [rootTrees]);

  const toggleExpand = (nodeId: string, namespaceId: string) => {
    const root = roots.find((r) => (r.rootNodeId || r.namespaceId) === nodeId);
    if (root?.rootNodeId && !rootTrees.has(root.rootNodeId)) {
      loadTreeForRoot(root.rootNodeId, namespaceId);
      return;
    }
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) next.delete(nodeId);
      else next.add(nodeId);
      return next;
    });
  };

  const visibleTree = React.useMemo((): TreeItem[] => {
    const result: TreeItem[] = [];
    for (const r of roots) {
      const rowId = r.rootNodeId || r.namespaceId;
      const stored = r.rootNodeId ? rootTrees.get(r.rootNodeId) : null;
      const treeNode = stored?.tree ?? null;
      const namespaceId = stored?.namespaceId ?? r.namespaceId;
      const expanded = expandedNodes.has(rowId);
      const children: Node[] = (treeNode?.children || []).filter((c): c is NonNullable<typeof c> => c != null).map((c) => ({ node_id: c.node_id, title: c.title }));
      const signals = r.rootNodeId ? getSignals(r.rootNodeId, treeNode) : undefined;
      const rootRoles = treeNode?.roles ?? [];
      result.push({
        node: { node_id: rowId, title: r.label },
        children,
        level: 0,
        expanded,
        namespaceId,
        expandable: Boolean(r.rootNodeId),
        rootNodeId: r.rootNodeId || undefined,
        nodeType: getNodeType(rootRoles),
        signals,
      });
      if (expanded && treeNode?.children) {
        function walk(t: ProductTreeVM['tree'], level: number, nsId: string) {
          if (!t) return;
          const exp = expandedNodes.has(t.node_id);
          const kids: Node[] = (t.children || []).filter((c): c is NonNullable<typeof c> => c != null).map((c) => ({ node_id: c.node_id, title: c.title }));
          const sig = getSignals(t.node_id, t);
          const nodeType = getNodeType(t.roles ?? []);
          result.push({
            node: { node_id: t.node_id, title: t.title },
            children: kids,
            level,
            expanded: exp,
            namespaceId: nsId,
            expandable: true,
            nodeType,
            signals: sig,
          });
          if (exp) for (const c of (t.children || []).filter((c): c is NonNullable<typeof c> => c != null)) walk(c, level + 1, nsId);
        }
        for (const c of treeNode.children) walk(c, 1, namespaceId);
      }
    }
    return result;
  }, [roots, rootTrees, expandedNodes]);

  const findPathInTree = useCallback((nodeId: string, tree: ProductTreeVM['tree'], path: string[]): boolean => {
    if (!tree) return false;
    path.push(tree.node_id);
    if (tree.node_id === nodeId) return true;
    for (const c of tree.children || []) {
      if (findPathInTree(nodeId, c, path)) return true;
    }
    path.pop();
    return false;
  }, []);

  useEffect(() => {
    if (!pendingNodeId || !onClearPendingNavigation || roots.length === 0) return;
    for (const r of roots) {
      const stored = rootTrees.get(r.rootNodeId);
      if (!stored?.tree) continue;
      const path: string[] = [];
      if (findPathInTree(pendingNodeId, stored.tree, path)) {
        setExpandedNodes((prev) => {
          const next = new Set(prev);
          path.forEach((id) => next.add(id));
          return next;
        });
        onSelectNode(pendingNodeId, r.namespaceId);
        onClearPendingNavigation();
        return;
      }
    }
    roots.forEach((r) => {
      if (!rootTrees.has(r.rootNodeId)) loadTreeForRoot(r.rootNodeId, r.namespaceId);
    });
  }, [pendingNodeId, roots, rootTrees, findPathInTree, loadTreeForRoot, onSelectNode, onClearPendingNavigation]);

  useEffect(() => {
    const handleNavigate = (event: Event) => {
      const customEvent = event as CustomEvent<{ nodeId: string }>;
      const { nodeId } = customEvent.detail;
      for (const r of roots) {
        const stored = rootTrees.get(r.rootNodeId);
        if (!stored?.tree) continue;
        const path: string[] = [];
        if (findPathInTree(nodeId, stored.tree, path)) {
          setExpandedNodes((prev) => {
            const next = new Set(prev);
            path.forEach((id) => next.add(id));
            return next;
          });
          onSelectNode(nodeId, r.namespaceId);
          break;
        }
      }
    };
    window.addEventListener('navigateToNode:products', handleNavigate);
    return () => window.removeEventListener('navigateToNode:products', handleNavigate);
  }, [roots, rootTrees, findPathInTree, onSelectNode]);

  return (
    <div style={{ height: '100%', overflow: 'auto', padding: '1rem' }}>
      {loadingRoots && <div className="loading">Loading ledgers...</div>}
      {error && <div className="error">Error: {error}</div>}
      {!loadingRoots && !error && roots.length === 0 && (
        <div className="loading">No ledger roots found</div>
      )}
      {!loadingRoots && !error && visibleTree.length > 0 && (
        <nav aria-label="Ledgers">
          <p style={{ fontSize: '0.75rem', color: '#666', marginBottom: '0.5rem' }}>
            Click a row or the <strong>▶</strong> button to expand and see the hierarchy.
          </p>
          <div style={{ marginTop: '0.5rem' }}>
            {visibleTree.map((item, idx) => (
                <div
                  key={`${item.node.node_id}-${item.level}-${idx}`}
                  title={item.node.title}
                  style={{
                    marginLeft: `${item.level * 20}px`,
                    padding: '0.5rem',
                    marginBottom: '0.25rem',
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: '0.5rem',
                    cursor: 'pointer',
                    background: (selectedNodeId === item.node.node_id) || (item.level === 0 && !item.rootNodeId && contextNamespaceId === item.namespaceId) ? '#e3f2fd' : 'transparent',
                    borderRadius: '4px',
                  }}
                  onClick={() => {
                    if (item.level === 0 && item.rootNodeId && !rootTrees.has(item.rootNodeId)) {
                      loadTreeForRoot(item.rootNodeId, item.namespaceId);
                    }
                    onSelectNode(item.level === 0 && item.rootNodeId === undefined ? '' : item.node.node_id, item.namespaceId);
                  }}
                >
                  {/* Always show expand control for namespace rows (level 0); for children show only when expandable and has children */}
                  {(item.level === 0) || (item.expandable && item.children.length > 0) ? (
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleExpand(item.node.node_id, item.namespaceId);
                    }}
                    disabled={item.level === 0 && !item.rootNodeId}
                    aria-label={item.expanded ? 'Collapse' : 'Expand'}
                    title={item.level === 0 && !item.rootNodeId ? 'No tree for this namespace' : item.expanded ? 'Collapse tree' : 'Expand tree'}
                    style={{
                      flexShrink: 0,
                      background: item.level === 0 && !item.rootNodeId ? '#eee' : '#e8e8e8',
                      border: '1px solid #999',
                      borderRadius: '4px',
                      cursor: item.level === 0 && !item.rootNodeId ? 'not-allowed' : 'pointer',
                      fontSize: '0.8rem',
                      width: '28px',
                      height: '28px',
                      padding: 0,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#222',
                      fontWeight: 'bold',
                    }}
                  >
                    {item.level === 0 && loadingRootId === item.node.node_id
                      ? '…'
                      : item.expanded
                        ? '▼'
                        : '▶'}
                  </button>
                ) : (
                  <span style={{ width: '28px', flexShrink: 0 }} />
                )}
                <div style={{ flex: 1, minWidth: 0, overflow: 'hidden' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                    {item.nodeType && (
                      <span
                        title={item.nodeType}
                        style={{
                          fontSize: '0.6875rem',
                          fontWeight: '600',
                          color: getNodeTypeColor(item.nodeType),
                          background: `${getNodeTypeColor(item.nodeType)}18`,
                          padding: '0.125rem 0.375rem',
                          borderRadius: '4px',
                          lineHeight: 1.2,
                        }}
                      >
                        {getNodeTypeShort(item.nodeType)}
                      </span>
                    )}
                    <div style={{ fontWeight: item.level === 0 ? 'bold' : 'normal', wordBreak: 'break-word', overflowWrap: 'break-word' }}>
                      {item.node.title}
                    </div>
                    {item.signals && (
                      <div style={{ display: 'flex', gap: '0.25rem', fontSize: '0.75rem' }}>
                        {item.signals.alignment && <span title="Has alignment links" style={{ color: '#4caf50' }}>●</span>}
                        {item.signals.coherence && <span title="Has coherence links" style={{ color: '#2196f3' }}>●</span>}
                        {item.signals.decision && <span title="Has decisions/evidence" style={{ color: '#ff9800' }}>●</span>}
                        {item.signals.materials && <span title="Has materials" style={{ color: '#9c27b0' }}>●</span>}
                      </div>
                    )}
                  </div>
                  {item.children.length > 0 && (
                    <div style={{ fontSize: '0.875rem', color: '#666', marginTop: '0.25rem' }}>
                      {item.children.length} child{item.children.length !== 1 ? 'ren' : ''}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </nav>
      )}
    </div>
  );
}
