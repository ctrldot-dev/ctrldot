import React, { useState, useEffect } from 'react';
import { guiClient } from '../adapters/GuiClient.js';
import ProductsTab from './ProductsTab.js';
import IntentTab from './IntentTab.js';
import LedgerTab from './LedgerTab.js';
import MaterialsTab from './MaterialsTab.js';
import MarkdownViewer from './MarkdownViewer.js';
import ConnectionIndicator from './ConnectionIndicator.js';

type Tab = 'products' | 'intent' | 'ledger' | 'materials';
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
  const [activeTab, setActiveTab] = useState<Tab>('products');
  const [connected, setConnected] = useState(false);
  const [toolTabs, setToolTabs] = useState<ToolTab[]>([]);
  const [activeToolTab, setActiveToolTab] = useState<string | null>(null);

  useEffect(() => {
    // Check connection on mount
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
    // Listen for openMaterial events from MaterialsTab
    const handleOpenMaterial = (event: Event) => {
      const customEvent = event as CustomEvent<{
        material_id: string;
        node_id: string;
        title: string;
        content_ref: string;
        media_type: string;
      }>;
      const material = customEvent.detail;
      
      // Check if tool tab already exists
      const existing = toolTabs.find(t => t.material.material_id === material.material_id);
      if (existing) {
        setActiveToolTab(existing.id);
      } else {
        const newTab: ToolTab = {
          id: `tool-${material.material_id}`,
          title: material.title,
          material,
        };
        setToolTabs([...toolTabs, newTab]);
        setActiveToolTab(newTab.id);
      }
    };

    // Listen for openInTab events from Details drawer
    const handleOpenInTab = (event: Event) => {
      const customEvent = event as CustomEvent<{
        tab: Tab;
        nodeId: string;
      }>;
      const { tab, nodeId } = customEvent.detail;
      setActiveTab(tab);
      // Dispatch to the specific tab component to handle navigation
      window.dispatchEvent(new CustomEvent(`navigateToNode:${tab}`, { detail: { nodeId } }));
    };

    window.addEventListener('openMaterial', handleOpenMaterial);
    window.addEventListener('openInTab', handleOpenInTab);
    return () => {
      window.removeEventListener('openMaterial', handleOpenMaterial);
      window.removeEventListener('openInTab', handleOpenInTab);
    };
  }, [toolTabs]);

  const closeToolTab = (tabId: string) => {
    setToolTabs(toolTabs.filter(t => t.id !== tabId));
    if (activeToolTab === tabId) {
      setActiveToolTab(toolTabs.length > 1 ? toolTabs.find(t => t.id !== tabId)?.id || null : null);
    }
  };

  const activeToolTabData = toolTabs.find(t => t.id === activeToolTab);

  return (
    <div id="app">
      <div className="header">
        <h1>Product Ledger Demo</h1>
        <ConnectionIndicator connected={connected} />
      </div>
      <div className="tabs">
        <button
          className={`tab ${activeTab === 'products' ? 'active' : ''}`}
          onClick={() => setActiveTab('products')}
        >
          Products
        </button>
        <button
          className={`tab ${activeTab === 'intent' ? 'active' : ''}`}
          onClick={() => setActiveTab('intent')}
        >
          Intent
        </button>
        <button
          className={`tab ${activeTab === 'ledger' ? 'active' : ''}`}
          onClick={() => setActiveTab('ledger')}
        >
          Ledger
        </button>
        <button
          className={`tab ${activeTab === 'materials' ? 'active' : ''}`}
          onClick={() => setActiveTab('materials')}
        >
          Materials
        </button>
      </div>
      <div style={{ display: 'flex', height: 'calc(100vh - 120px)' }}>
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          <div className="content" style={{ flex: 1 }}>
            {activeTab === 'products' && <ProductsTab />}
            {activeTab === 'intent' && <IntentTab />}
            {activeTab === 'ledger' && <LedgerTab />}
            {activeTab === 'materials' && <MaterialsTab />}
          </div>
        </div>
        {toolTabs.length > 0 && (
          <div style={{ width: '600px', borderLeft: '1px solid #e0e0e0', display: 'flex', flexDirection: 'column' }}>
            <div style={{ display: 'flex', borderBottom: '1px solid #e0e0e0', overflowX: 'auto' }}>
              {toolTabs.map((tab) => (
                <div
                  key={tab.id}
                  style={{
                    padding: '0.5rem 1rem',
                    borderRight: '1px solid #e0e0e0',
                    background: activeToolTab === tab.id ? '#e3f2fd' : 'white',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    minWidth: '150px',
                  }}
                  onClick={() => setActiveToolTab(tab.id)}
                >
                  <span style={{ flex: 1, fontSize: '0.875rem' }}>{tab.title}</span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      closeToolTab(tab.id);
                    }}
                    style={{
                      background: 'transparent',
                      border: 'none',
                      cursor: 'pointer',
                      fontSize: '1.2rem',
                      lineHeight: 1,
                      padding: 0,
                      width: '20px',
                      height: '20px',
                    }}
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
            {activeToolTabData && (
              <div style={{ flex: 1, overflow: 'hidden' }}>
                <MarkdownViewer
                  material={activeToolTabData.material}
                  onClose={() => closeToolTab(activeToolTabData.id)}
                />
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
