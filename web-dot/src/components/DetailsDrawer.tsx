import React from 'react';
import type { NodeVM } from '../adapters/GuiClient.js';

interface DetailsDrawerProps {
  nodeData: NodeVM | null;
  onClose: () => void;
  onOpenInTab?: (tab: 'products' | 'intent' | 'ledger' | 'materials', nodeId: string) => void;
}

export default function DetailsDrawer({ nodeData, onClose, onOpenInTab }: DetailsDrawerProps) {
  if (!nodeData) return null;

  const { node, relationships } = nodeData;

  return (
    <div
      style={{
        width: '400px',
        borderLeft: '1px solid #e0e0e0',
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        background: 'white',
      }}
    >
      <div
        style={{
          padding: '1rem',
          borderBottom: '1px solid #e0e0e0',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <h2 style={{ margin: 0, fontSize: '1.25rem' }}>Details</h2>
        <button
          onClick={onClose}
          style={{
            background: 'transparent',
            border: 'none',
            cursor: 'pointer',
            fontSize: '1.5rem',
            lineHeight: 1,
            padding: 0,
            width: '30px',
            height: '30px',
          }}
        >
          ×
        </button>
      </div>

      <div style={{ flex: 1, overflow: 'auto', padding: '1rem' }}>
        <div style={{ marginBottom: '1.5rem' }}>
          <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1rem', fontWeight: '600' }}>{node.title}</h3>
          <div style={{ fontSize: '0.875rem', color: '#666' }}>
            <div>ID: {node.node_id}</div>
            {node.roles.length > 0 && (
              <div style={{ marginTop: '0.25rem' }}>
                Roles: {node.roles.join(', ')}
              </div>
            )}
          </div>
          {node.description && (
            <div style={{ marginTop: '1rem', padding: '0.75rem', background: '#f5f5f5', borderRadius: '4px' }}>
              <div style={{ fontSize: '0.875rem', fontWeight: '600', marginBottom: '0.5rem' }}>Description</div>
              <div style={{ fontSize: '0.875rem', lineHeight: '1.5' }}>{node.description}</div>
            </div>
          )}
          {node.jtbd && (
            <div style={{ marginTop: '1rem', padding: '0.75rem', background: '#e3f2fd', borderRadius: '4px' }}>
              <div style={{ fontSize: '0.875rem', fontWeight: '600', marginBottom: '0.5rem' }}>Jobs to be Done</div>
              <div style={{ fontSize: '0.875rem', lineHeight: '1.5' }}>{node.jtbd}</div>
            </div>
          )}
        </div>

        {/* Children */}
        {relationships.children.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Children ({relationships.children.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {relationships.children.map((child) => (
                <div
                  key={child.node_id}
                  onClick={() => onOpenInTab?.('products', child.node_id)}
                  style={{
                    padding: '0.5rem',
                    border: '1px solid #e0e0e0',
                    borderRadius: '4px',
                    cursor: onOpenInTab ? 'pointer' : 'default',
                    background: onOpenInTab ? '#f5f5f5' : 'white',
                  }}
                >
                  <div style={{ fontWeight: '500', fontSize: '0.875rem' }}>{child.title}</div>
                  {child.roles.length > 0 && (
                    <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
                      {child.roles.join(', ')}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Alignment */}
        {relationships.alignment.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Alignment ({relationships.alignment.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {relationships.alignment.map((rel, idx) => (
                <div
                  key={`${rel.node_id}-${idx}`}
                  onClick={() => onOpenInTab?.('products', rel.node_id)}
                  style={{
                    padding: '0.5rem',
                    border: '1px solid #e0e0e0',
                    borderRadius: '4px',
                    cursor: onOpenInTab ? 'pointer' : 'default',
                    background: onOpenInTab ? '#f5f5f5' : 'white',
                  }}
                >
                  <div style={{ fontWeight: '500', fontSize: '0.875rem' }}>{rel.title}</div>
                  <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
                    {rel.type}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Coherence */}
        {relationships.coherence.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Coherence ({relationships.coherence.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {relationships.coherence.map((rel, idx) => (
                <div
                  key={`${rel.node_id}-${idx}`}
                  onClick={() => onOpenInTab?.('products', rel.node_id)}
                  style={{
                    padding: '0.5rem',
                    border: '1px solid #e0e0e0',
                    borderRadius: '4px',
                    cursor: onOpenInTab ? 'pointer' : 'default',
                    background: onOpenInTab ? '#f5f5f5' : 'white',
                  }}
                >
                  <div style={{ fontWeight: '500', fontSize: '0.875rem' }}>{rel.title}</div>
                  <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
                    {rel.type}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Decisions & Evidence */}
        {relationships.decisions_and_evidence.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Decisions & Evidence ({relationships.decisions_and_evidence.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {relationships.decisions_and_evidence.map((rel, idx) => (
                <div
                  key={`${rel.node_id}-${idx}`}
                  onClick={() => onOpenInTab?.('products', rel.node_id)}
                  style={{
                    padding: '0.5rem',
                    border: '1px solid #e0e0e0',
                    borderRadius: '4px',
                    cursor: onOpenInTab ? 'pointer' : 'default',
                    background: onOpenInTab ? '#f5f5f5' : 'white',
                  }}
                >
                  <div style={{ fontWeight: '500', fontSize: '0.875rem' }}>{rel.title}</div>
                  <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
                    {rel.type}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Materials */}
        {relationships.materials.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Materials ({relationships.materials.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {relationships.materials.map((material) => (
                <div
                  key={material.material_id}
                  onClick={() => {
                    window.dispatchEvent(new CustomEvent('openMaterial', {
                      detail: {
                        material_id: material.material_id,
                        node_id: node.node_id,
                        title: (material.meta?.title as string) || material.content_ref.split('/').pop() || material.material_id,
                        content_ref: material.content_ref,
                        media_type: material.media_type,
                      }
                    }));
                    onOpenInTab?.('materials', node.node_id);
                  }}
                  style={{
                    padding: '0.5rem',
                    border: '1px solid #e0e0e0',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    background: '#f5f5f5',
                  }}
                >
                  <div style={{ fontWeight: '500', fontSize: '0.875rem' }}>
                    {(material.meta?.title as string) || material.content_ref.split('/').pop() || material.material_id}
                  </div>
                  <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
                    {material.media_type}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Actions */}
        {onOpenInTab && (
          <div style={{ marginTop: '1.5rem', paddingTop: '1.5rem', borderTop: '1px solid #e0e0e0' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Actions
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              <button
                onClick={() => onOpenInTab('products', node.node_id)}
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
                onClick={() => onOpenInTab('intent', node.node_id)}
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
                onClick={() => onOpenInTab('ledger', node.node_id)}
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
              <button
                onClick={() => onOpenInTab('materials', node.node_id)}
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
                Open in Materials
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
