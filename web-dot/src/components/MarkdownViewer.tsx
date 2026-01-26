import React, { useState, useEffect } from 'react';
import { guiClient, type NodeVM } from '../adapters/GuiClient.js';

interface MarkdownViewerProps {
  material: {
    material_id: string;
    node_id: string;
    title: string;
    content_ref: string;
    media_type: string;
  };
  onClose: () => void;
}

export default function MarkdownViewer({ material, onClose }: MarkdownViewerProps) {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [nodeData, setNodeData] = useState<NodeVM | null>(null);
  const [loadingNode, setLoadingNode] = useState(false);

  useEffect(() => {
    loadContent();
    loadNodeData();
  }, [material.content_ref, material.node_id]);

  const loadNodeData = async () => {
    setLoadingNode(true);
    try {
      const data = await guiClient.node(material.node_id);
      setNodeData(data);
    } catch (err) {
      console.error('Error loading node data:', err);
      setNodeData(null);
    } finally {
      setLoadingNode(false);
    }
  };

  const handleOpenInTab = (tab: 'products' | 'intent' | 'ledger' | 'materials', nodeId: string) => {
    window.dispatchEvent(new CustomEvent('openInTab', {
      detail: { tab, nodeId }
    }));
  };

  const loadContent = async () => {
    setLoading(true);
    setError(null);
    try {
      // For now, assume content_ref is a URL or path
      // In production, this would fetch from the kernel's material storage
      const response = await fetch(material.content_ref);
      if (!response.ok) {
        throw new Error(`Failed to load material: ${response.statusText}`);
      }
      const text = await response.text();
      setContent(text);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load content');
      console.error('Error loading material content:', err);
    } finally {
      setLoading(false);
    }
  };

  // Simple markdown rendering (basic implementation)
  // In production, use a proper markdown library like marked or markdown-it
  const renderMarkdown = (md: string) => {
    // Very basic markdown rendering for demo
    let html = md
      .replace(/^# (.*$)/gim, '<h1>$1</h1>')
      .replace(/^## (.*$)/gim, '<h2>$1</h2>')
      .replace(/^### (.*$)/gim, '<h3>$1</h3>')
      .replace(/^\*\*(.*)\*\*/gim, '<strong>$1</strong>')
      .replace(/^\*(.*)\*/gim, '<em>$1</em>')
      .replace(/\n\n/g, '</p><p>')
      .replace(/\n/g, '<br>');
    
    return `<p>${html}</p>`;
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ padding: '1rem', borderBottom: '1px solid #e0e0e0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>{material.title}</h2>
        <button
          onClick={onClose}
          style={{
            padding: '0.5rem 1rem',
            background: '#666',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Close
        </button>
      </div>
      <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        <div style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
          {loading && <div className="loading">Loading content...</div>}
          {error && <div className="error">Error: {error}</div>}
          {!loading && !error && (
            <div
              style={{
                maxWidth: '800px',
                margin: '0 auto',
                lineHeight: '1.6',
              }}
              dangerouslySetInnerHTML={{ __html: renderMarkdown(content) }}
            />
          )}
        </div>
        {/* Material Context Panel */}
        <div
          style={{
            width: '300px',
            borderLeft: '1px solid #e0e0e0',
            padding: '1rem',
            overflow: 'auto',
            background: '#f9f9f9',
          }}
        >
          <h3 style={{ margin: '0 0 1rem 0', fontSize: '1rem', fontWeight: '600' }}>Context</h3>
          {loadingNode && <div className="loading">Loading node...</div>}
          {!loadingNode && nodeData && (
            <div>
              <div style={{ marginBottom: '1rem', padding: '0.75rem', background: 'white', borderRadius: '4px', border: '1px solid #e0e0e0' }}>
                <div style={{ fontSize: '0.875rem', fontWeight: '600', marginBottom: '0.5rem' }}>Linked Node</div>
                <div style={{ fontSize: '0.875rem', marginBottom: '0.25rem' }}>{nodeData.node.title}</div>
                <div style={{ fontSize: '0.75rem', color: '#666', fontFamily: 'monospace' }}>{nodeData.node.node_id}</div>
                {nodeData.node.roles.length > 0 && (
                  <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
                    Roles: {nodeData.node.roles.join(', ')}
                  </div>
                )}
              </div>
              <div style={{ marginBottom: '1rem' }}>
                <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
                  Actions
                </h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  <button
                    onClick={() => handleOpenInTab('products', material.node_id)}
                    style={{
                      padding: '0.5rem',
                      border: '1px solid #e0e0e0',
                      borderRadius: '4px',
                      background: 'white',
                      cursor: 'pointer',
                      textAlign: 'left',
                      fontSize: '0.875rem',
                    }}
                  >
                    Open in Products
                  </button>
                  <button
                    onClick={() => handleOpenInTab('intent', material.node_id)}
                    style={{
                      padding: '0.5rem',
                      border: '1px solid #e0e0e0',
                      borderRadius: '4px',
                      background: 'white',
                      cursor: 'pointer',
                      textAlign: 'left',
                      fontSize: '0.875rem',
                    }}
                  >
                    Open in Intent
                  </button>
                  <button
                    onClick={() => handleOpenInTab('ledger', material.node_id)}
                    style={{
                      padding: '0.5rem',
                      border: '1px solid #e0e0e0',
                      borderRadius: '4px',
                      background: 'white',
                      cursor: 'pointer',
                      textAlign: 'left',
                      fontSize: '0.875rem',
                    }}
                  >
                    Open in Ledger
                  </button>
                </div>
              </div>
              {nodeData.relationships.materials.length > 1 && (
                <div>
                  <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
                    Other Materials ({nodeData.relationships.materials.length - 1})
                  </h4>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    {nodeData.relationships.materials
                      .filter(m => m.material_id !== material.material_id)
                      .map((m) => (
                        <div
                          key={m.material_id}
                          onClick={() => {
                            window.dispatchEvent(new CustomEvent('openMaterial', {
                              detail: {
                                material_id: m.material_id,
                                node_id: material.node_id,
                                title: (m.meta?.title as string) || m.content_ref.split('/').pop() || m.material_id,
                                content_ref: m.content_ref,
                                media_type: m.media_type,
                              }
                            }));
                          }}
                          style={{
                            padding: '0.5rem',
                            border: '1px solid #e0e0e0',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            background: 'white',
                            fontSize: '0.875rem',
                          }}
                        >
                          {(m.meta?.title as string) || m.content_ref.split('/').pop() || m.material_id}
                        </div>
                      ))}
                  </div>
                </div>
              )}
            </div>
          )}
          {!loadingNode && !nodeData && (
            <div style={{ fontSize: '0.875rem', color: '#666' }}>Node not found</div>
          )}
        </div>
      </div>
    </div>
  );
}
