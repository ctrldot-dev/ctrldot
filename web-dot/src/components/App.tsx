import React, { useState, useEffect } from 'react';
import { guiClient } from '../adapters/GuiClient.js';
import type { NodeVM } from '../adapters/GuiClient.js';
import LedgerTreeNav from './LedgerTreeNav.js';
import IntentTab from './IntentTab.js';
import LedgerTab from './LedgerTab.js';
import MaterialsTab from './MaterialsTab.js';
import ExplorerTab from './ExplorerTab.js';
import MarkdownViewer from './MarkdownViewer.js';
import ConnectionIndicator from './ConnectionIndicator.js';
import DetailsDrawer from './DetailsDrawer.js';

type ViewTab = 'intent' | 'ledger' | 'materials' | 'explorer';
type ToolTab = {
  id: string;
  title: string;
  material: {
    material_id: string;
    node_id: string;
    title: string;
    content_ref: string;
    media_type: string;
  };
};

export default function App() {
  const [activeTab, setActiveTab] = useState<ViewTab>('intent');
  const [connected, setConnected] = useState(false);
  const [toolTabs, setToolTabs] = useState<ToolTab[]>([]);
  const [activeToolTab, setActiveToolTab] = useState<string | null>(null);
  const [pendingProductsNodeId, setPendingProductsNodeId] = useState<string | null>(null);

  // Navigation context: the node selected in the left nav; drives Intent / Ledger / Materials
  const [contextNodeId, setContextNodeId] = useState<string | null>(null);
  const [contextNamespaceId, setContextNamespaceId] = useState<string | null>(null);
  // Inspector: the node whose details are shown in the Details pane (passive)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [selectedNamespaceId, setSelectedNamespaceId] = useState<string | null>(null);
  const [nodeDetails, setNodeDetails] = useState<NodeVM | null>(null);
  const [loadingDetails, setLoadingDetails] = useState(false);

  const handleSelectNode = (nodeId: string, namespaceId: string) => {
    const nodeIdOrNull = nodeId || null;
    setContextNodeId(nodeIdOrNull);
    setContextNamespaceId(namespaceId);
    setSelectedNodeId(nodeIdOrNull);
    setSelectedNamespaceId(namespaceId);
    setNodeDetails(null);
    if (nodeIdOrNull) {
      setLoadingDetails(true);
      guiClient
        .node(nodeIdOrNull, namespaceId)
        .then(setNodeDetails)
        .catch(() => setNodeDetails(null))
        .finally(() => setLoadingDetails(false));
    } else {
      setLoadingDetails(false);
    }
  };

  useEffect(() => {
    const checkConnection = async () => {
      try {
        const result = await guiClient.healthz();
        setConnected(result.ok);
      } catch {
        setConnected(false);
      }
    };
    checkConnection();
    const interval = setInterval(checkConnection, 5000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const handleOpenMaterial = (event: Event) => {
      const customEvent = event as CustomEvent<{
        material_id: string;
        node_id: string;
        title: string;
        content_ref: string;
        media_type: string;
      }>;
      const material = customEvent.detail;
      const existing = toolTabs.find((t) => t.material.material_id === material.material_id);
      if (existing) {
        setActiveToolTab(existing.id);
      } else {
        const newTab: ToolTab = {
          id: `tool-${material.material_id}`,
          title: material.title,
          material,
        };
        setToolTabs((prev) => [...prev, newTab]);
        setActiveToolTab(newTab.id);
      }
    };

    const handleOpenInTab = (event: Event) => {
      const customEvent = event as CustomEvent<{
        tab: ViewTab | 'products';
        nodeId: string;
      }>;
      const { tab, nodeId } = customEvent.detail;
      if (tab === 'products') {
        setPendingProductsNodeId(nodeId);
        window.dispatchEvent(new CustomEvent('navigateToNode:products', { detail: { nodeId } }));
      } else {
        setActiveTab(tab);
        setActiveToolTab(null);
        window.dispatchEvent(new CustomEvent(`navigateToNode:${tab}`, { detail: { nodeId } }));
      }
    };

    window.addEventListener('openMaterial', handleOpenMaterial);
    window.addEventListener('openInTab', handleOpenInTab);
    return () => {
      window.removeEventListener('openMaterial', handleOpenMaterial);
      window.removeEventListener('openInTab', handleOpenInTab);
    };
  }, [toolTabs]);

  const closeToolTab = (tabId: string) => {
    setToolTabs((prev) => prev.filter((t) => t.id !== tabId));
    if (activeToolTab === tabId) {
      setActiveToolTab(toolTabs.length > 1 ? toolTabs.find((t) => t.id !== tabId)?.id ?? null : null);
    }
  };

  const activeToolTabData = toolTabs.find((t) => t.id === activeToolTab);

  const handleCloseDrawer = () => {
    setSelectedNodeId(null);
    setSelectedNamespaceId(null);
    setNodeDetails(null);
  };
  const hasContext = contextNamespaceId != null;

  return (
    <div id="app">
      <div className="header">
        <h1>Decision Ledger Demo</h1>
        <ConnectionIndicator connected={connected} />
      </div>
      <div style={{ display: 'flex', height: 'calc(100vh - 56px)' }}>
        <aside style={{ width: '360px', minWidth: '320px', flexShrink: 0, borderRight: '1px solid #e0e0e0' }}>
          <LedgerTreeNav
            selectedNodeId={selectedNodeId}
            contextNamespaceId={contextNamespaceId}
            onSelectNode={handleSelectNode}
            pendingNodeId={pendingProductsNodeId}
            onClearPendingNavigation={() => setPendingProductsNodeId(null)}
          />
        </aside>
        <main style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
          <div className="tabs">
            <button
              className={`tab ${activeTab === 'intent' ? 'active' : ''}`}
              onClick={() => { setActiveTab('intent'); setActiveToolTab(null); }}
            >
              Intent
            </button>
            <button
              className={`tab ${activeTab === 'ledger' ? 'active' : ''}`}
              onClick={() => { setActiveTab('ledger'); setActiveToolTab(null); }}
            >
              Ledger
            </button>
            <button
              className={`tab ${activeTab === 'materials' ? 'active' : ''}`}
              onClick={() => { setActiveTab('materials'); setActiveToolTab(null); }}
            >
              Materials
            </button>
            <button
              className={`tab ${activeTab === 'explorer' ? 'active' : ''}`}
              onClick={() => { setActiveTab('explorer'); setActiveToolTab(null); }}
            >
              Explorer
            </button>
            {toolTabs.map((tab) => (
              <button
                key={tab.id}
                className={`tab ${activeToolTab === tab.id ? 'active' : ''}`}
                onClick={() => setActiveToolTab(tab.id)}
                style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
              >
                <span style={{ flex: 1, fontSize: '0.875rem' }}>{tab.title}</span>
                <span
                  onClick={(e) => { e.stopPropagation(); closeToolTab(tab.id); }}
                  style={{ fontSize: '1rem', lineHeight: 1, paddingLeft: '0.25rem', cursor: 'pointer' }}
                >
                  ×
                </span>
              </button>
            ))}
          </div>
          <div className="content" style={{ flex: 1, overflow: 'auto' }}>
            {!hasContext && !activeToolTab && activeTab !== 'explorer' && (
              <div style={{ padding: '2rem', color: '#666', textAlign: 'center' }}>
                <p style={{ fontSize: '1.125rem', marginBottom: '0.5rem' }}>
                  Select a node in the left nav to view Intent, Ledger, and Materials for that context.
                </p>
                <p style={{ fontSize: '0.875rem' }}>
                  The left nav shows all root ledgers (Product, Treasury, Stablecoin, etc.). Click any node to set it as the active context and re-root the main content.
                </p>
              </div>
            )}
            {hasContext && !activeToolTab && activeTab === 'intent' && (
              <IntentTab contextNamespaceId={contextNamespaceId} />
            )}
            {hasContext && !activeToolTab && activeTab === 'ledger' && (
              <LedgerTab contextNamespaceId={contextNamespaceId} />
            )}
            {hasContext && !activeToolTab && activeTab === 'materials' && (
              <MaterialsTab contextNodeId={contextNodeId} contextNamespaceId={contextNamespaceId} />
            )}
            {!activeToolTab && activeTab === 'explorer' && (
              <ExplorerTab contextNodeId={contextNodeId} contextNamespaceId={contextNamespaceId} />
            )}
            {activeToolTab && activeToolTabData && (
              <MarkdownViewer
                material={activeToolTabData.material}
                onClose={() => closeToolTab(activeToolTabData!.id)}
              />
            )}
          </div>
        </main>
        {selectedNodeId && (
          <>
            {loadingDetails && (
              <div
                style={{
                  width: '400px',
                  minWidth: '400px',
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
            {!loadingDetails && (
              <DetailsDrawer
                nodeData={nodeDetails}
                onClose={handleCloseDrawer}
                onOpenInTab={(tab, nodeId) =>
                  window.dispatchEvent(new CustomEvent('openInTab', { detail: { tab, nodeId } }))
                }
              />
            )}
          </>
        )}
      </div>
    </div>
  );
}
