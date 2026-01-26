import React, { useState, useEffect } from 'react';
import { guiClient, type NodeVM } from '../adapters/GuiClient.js';
import type { Node } from '../types/domain.js';
import DetailsDrawer from './DetailsDrawer.js';

export default function ProductsTab() {
  const [selectedProduct, setSelectedProduct] = useState<string | null>(null);
  const [tree, setTree] = useState<Array<{ node: Node; children: Node[]; level: number; expanded: boolean; signals?: { alignment: boolean; coherence: boolean; decision: boolean; materials: boolean } }>>([]);
  const [treeData, setTreeData] = useState<any>(null); // Store full tree data for signals
  const [root, setRoot] = useState<Node | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pinned, setPinned] = useState<{ fieldServe: string; assetLink: string } | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [nodeDetails, setNodeDetails] = useState<NodeVM | null>(null);
  const [loadingDetails, setLoadingDetails] = useState(false);

  useEffect(() => {
    (async () => {
      const cfg = await guiClient.config();
      setPinned(cfg.pinned_roots.products);
      // Default to FieldServe
      if (!selectedProduct) {
        loadProduct('FieldServe', cfg.pinned_roots.products.fieldServe);
      }
    })().catch((e) => setError(e instanceof Error ? e.message : 'Failed to load config'));
  }, []);

  const loadProduct = async (productKey: 'FieldServe' | 'AssetLink', rootId: string) => {
    setLoading(true);
    setError(null);
    try {
      const vm = await guiClient.productsTree(rootId, 10);
      if (!vm.tree) {
        setError('Tree not available (root not found).');
        return;
      }

      // Flatten to our tree rendering model with expand/collapse
      const flat: Array<{ node: Node; children: Node[]; level: number; expanded: boolean }> = [];

      function walk(t: any, level: number, parentExpanded: boolean = true) {
        const node: Node = { node_id: t.node_id, title: t.title };
        const children: Node[] = (t.children || []).map((c: any) => ({ node_id: c.node_id, title: c.title }));
        const expanded = level === 0 ? true : parentExpanded; // Root always expanded
        flat.push({ node, children, level, expanded });
        if (expanded) {
          for (const c of t.children || []) walk(c, level + 1, expanded);
        }
      }

      walk(vm.tree, 0);
      setRoot({ node_id: vm.tree.node_id, title: vm.tree.title });
      setTree(flat);
      setTreeData(vm.tree); // Store full tree for signals lookup
      setSelectedProduct(productKey);
      setSelectedNodeId(null);
      setNodeDetails(null);
      setExpandedNodes(new Set([vm.tree.node_id])); // Root expanded by default
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load product');
      console.error('Error loading product:', err);
    } finally {
      setLoading(false);
    }
  };

  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());

  const toggleExpand = (nodeId: string) => {
    setExpandedNodes(prev => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  // Helper to find signals for a node in treeData
  const getSignals = (nodeId: string, treeNode: any): { alignment: boolean; coherence: boolean; decision: boolean; materials: boolean } | undefined => {
    if (!treeNode) return undefined;
    if (treeNode.node_id === nodeId) return treeNode.signals;
    for (const child of treeNode.children || []) {
      const found = getSignals(nodeId, child);
      if (found) return found;
    }
    return undefined;
  };

  // Rebuild visible tree based on expanded state
  const visibleTree = React.useMemo(() => {
    if (!root) return [];
    
    const result: Array<{ node: Node; children: Node[]; level: number; expanded: boolean; signals?: { alignment: boolean; coherence: boolean; decision: boolean; materials: boolean } }> = [];
    const nodeMap = new Map(tree.map(item => [item.node.node_id, item]));
    
    function buildVisible(nodeId: string, level: number): void {
      const item = nodeMap.get(nodeId);
      if (!item) return;
      
      const expanded = level === 0 ? true : expandedNodes.has(nodeId);
      const signals = treeData ? getSignals(nodeId, treeData) : undefined;
      result.push({ ...item, expanded, signals });
      
      if (expanded) {
        for (const child of item.children) {
          buildVisible(child.node_id, level + 1);
        }
      }
    }
    
    buildVisible(root.node_id, 0);
    return result;
  }, [tree, root, expandedNodes, treeData]);

  const handleNodeClick = async (nodeId: string) => {
    setSelectedNodeId(nodeId);
    setLoadingDetails(true);
    try {
      const details = await guiClient.node(nodeId);
      setNodeDetails(details);
    } catch (err) {
      console.error('Error loading node details:', err);
      setNodeDetails(null);
    } finally {
      setLoadingDetails(false);
    }
  };

  const handleOpenInTab = (tab: 'products' | 'intent' | 'ledger' | 'materials', nodeId: string) => {
    window.dispatchEvent(new CustomEvent('openInTab', {
      detail: { tab, nodeId }
    }));
  };

  useEffect(() => {
    const handleNavigate = (event: Event) => {
      const customEvent = event as CustomEvent<{ nodeId: string }>;
      const { nodeId } = customEvent.detail;
      handleNodeClick(nodeId);
      // Expand path to node
      const path: string[] = [];
      function findPath(currentId: string, targetId: string, visited: Set<string>): boolean {
        if (visited.has(currentId)) return false;
        visited.add(currentId);
        if (currentId === targetId) {
          path.push(currentId);
          return true;
        }
        const item = tree.find(i => i.node.node_id === currentId);
        if (!item) return false;
        for (const child of item.children) {
          if (findPath(child.node_id, targetId, visited)) {
            path.push(currentId);
            return true;
          }
        }
        return false;
      }
      if (root) {
        findPath(root.node_id, nodeId, new Set());
        setExpandedNodes(prev => {
          const next = new Set(prev);
          for (const id of path) next.add(id);
          return next;
        });
      }
    };
    window.addEventListener('navigateToNode:products', handleNavigate);
    return () => window.removeEventListener('navigateToNode:products', handleNavigate);
  }, [tree, root]);

  return (
    <div style={{ display: 'flex', height: '100%', gap: 0 }}>
      <div style={{ width: '200px', borderRight: '1px solid #e0e0e0', padding: '1rem' }}>
        <h3 style={{ marginBottom: '1rem' }}>Products</h3>
        <button
          onClick={() => pinned && loadProduct('FieldServe', pinned.fieldServe)}
          style={{
            display: 'block',
            width: '100%',
            padding: '0.5rem',
            marginBottom: '0.5rem',
            textAlign: 'left',
            background: selectedProduct === 'FieldServe' ? '#e3f2fd' : 'transparent',
            border: '1px solid #e0e0e0',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          FieldServe
        </button>
        <button
          onClick={() => pinned && loadProduct('AssetLink', pinned.assetLink)}
          style={{
            display: 'block',
            width: '100%',
            padding: '0.5rem',
            marginBottom: '0.5rem',
            textAlign: 'left',
            background: selectedProduct === 'AssetLink' ? '#e3f2fd' : 'transparent',
            border: '1px solid #e0e0e0',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          AssetLink
        </button>
      </div>
      <div style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        {loading && <div className="loading">Loading product tree...</div>}
        {error && <div className="error">Error: {error}</div>}
        {!loading && !error && tree.length === 0 && (
          <div className="loading">No product selected or tree is empty</div>
        )}
        {!loading && !error && tree.length > 0 && (
          <div>
            <h2>{root?.title || 'Product Tree'}</h2>
            <div style={{ marginTop: '1rem' }}>
              {visibleTree.map((item, idx) => (
                <div
                  key={`${item.node.node_id}-${idx}`}
                  style={{
                    marginLeft: `${item.level * 20}px`,
                    padding: '0.5rem',
                    marginBottom: '0.25rem',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    cursor: 'pointer',
                    background: selectedNodeId === item.node.node_id ? '#e3f2fd' : 'transparent',
                    borderRadius: '4px',
                  }}
                  onClick={() => handleNodeClick(item.node.node_id)}
                >
                  {item.children.length > 0 && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleExpand(item.node.node_id);
                      }}
                      style={{
                        background: 'transparent',
                        border: 'none',
                        cursor: 'pointer',
                        fontSize: '0.875rem',
                        width: '20px',
                        padding: 0,
                      }}
                    >
                      {item.expanded ? '▼' : '▶'}
                    </button>
                  )}
                  {item.children.length === 0 && <span style={{ width: '20px' }} />}
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <div style={{ fontWeight: item.level === 0 ? 'bold' : 'normal' }}>
                        {item.node.title}
                      </div>
                      {/* Node signals */}
                      {item.signals && (
                        <div style={{ display: 'flex', gap: '0.25rem', fontSize: '0.75rem' }}>
                          {item.signals.alignment && (
                            <span title="Has alignment links" style={{ color: '#4caf50' }}>●</span>
                          )}
                          {item.signals.coherence && (
                            <span title="Has coherence links" style={{ color: '#2196f3' }}>●</span>
                          )}
                          {item.signals.decision && (
                            <span title="Has decisions/evidence" style={{ color: '#ff9800' }}>●</span>
                          )}
                          {item.signals.materials && (
                            <span title="Has materials" style={{ color: '#9c27b0' }}>●</span>
                          )}
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
          </div>
        )}
      </div>
      {selectedNodeId && (
        <DetailsDrawer
          nodeData={nodeDetails}
          onClose={() => {
            setSelectedNodeId(null);
            setNodeDetails(null);
          }}
          onOpenInTab={handleOpenInTab}
        />
      )}
      {loadingDetails && selectedNodeId && (
        <div
          style={{
            width: '400px',
            borderLeft: '1px solid #e0e0e0',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '2rem',
          }}
        >
          <div className="loading">Loading details...</div>
        </div>
      )}
    </div>
  );
}
