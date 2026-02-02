import React, { useState, useEffect } from 'react';
import type { NodeVM } from '../adapters/GuiClient.js';
import { getNodeType, getNodeTypeColor, type NodeTypeLabel } from '../utils/nodeType.js';

const TRUNCATE_LEN = 380;
function truncateMarkdown(md: string): string {
  const trimmed = md.trim();
  if (trimmed.length <= TRUNCATE_LEN) return trimmed;
  const at = trimmed.slice(0, TRUNCATE_LEN).lastIndexOf(' ');
  return (at > TRUNCATE_LEN / 2 ? trimmed.slice(0, at) : trimmed.slice(0, TRUNCATE_LEN)) + '…';
}

function renderMarkdownSnippet(md: string): string {
  let html = md
    .replace(/^# (.*)$/gim, '<strong>$1</strong>')
    .replace(/^## (.*)$/gim, '<strong>$1</strong>')
    .replace(/^### (.*)$/gim, '<strong>$1</strong>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n\n/g, '</p><p>')
    .replace(/\n/g, ' ');
  return `<p>${html}</p>`;
}

type MaterialEntry = { material_id: string; content_ref: string; media_type: string; meta: Record<string, unknown> };

function InlineMaterialBlock({
  material,
  nodeId,
  sectionLabel,
}: {
  material: MaterialEntry;
  nodeId: string;
  sectionLabel: string;
}) {
  const [content, setContent] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const url = material.content_ref.startsWith('/') || material.content_ref.startsWith('http') ? material.content_ref : `/materials/${material.material_id}.md`;

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetch(url)
      .then((r) => (r.ok ? r.text() : ''))
      .then((text) => { if (!cancelled) setContent(text || ''); })
      .catch(() => { if (!cancelled) setContent(''); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [url]);

  const openMaterial = () => {
    window.dispatchEvent(new CustomEvent('openMaterial', {
      detail: {
        material_id: material.material_id,
        node_id: nodeId,
        title: (material.meta?.title as string) || material.content_ref.split('/').pop() || material.material_id,
        content_ref: material.content_ref,
        media_type: material.media_type,
      },
    }));
  };

  const display = content != null ? truncateMarkdown(content) : '';

  return (
    <div style={{ marginTop: '1rem', padding: '0.75rem', background: '#f5f5f5', borderRadius: '4px' }}>
      <div style={{ fontSize: '0.875rem', fontWeight: '600', marginBottom: '0.5rem' }}>{sectionLabel}</div>
      {loading && <div style={{ fontSize: '0.875rem', color: '#888' }}>Loading…</div>}
      {!loading && (
        <>
          <div
            style={{ fontSize: '0.875rem', lineHeight: '1.5' }}
            dangerouslySetInnerHTML={{ __html: renderMarkdownSnippet(display) }}
          />
          <button
            type="button"
            onClick={openMaterial}
            style={{
              marginTop: '0.5rem',
              fontSize: '0.8125rem',
              background: 'transparent',
              border: 'none',
              color: '#1565c0',
              cursor: 'pointer',
              padding: 0,
              textDecoration: 'underline',
            }}
          >
            Open material
          </button>
        </>
      )}
    </div>
  );
}

interface DetailsDrawerProps {
  nodeData: NodeVM | null;
  onClose: () => void;
  onOpenInTab?: (tab: 'products' | 'intent' | 'ledger' | 'materials', nodeId: string) => void;
}

function matchMaterial(materials: MaterialEntry[], typeOrCategory: string): MaterialEntry | undefined {
  const t = (typeOrCategory || '').toLowerCase();
  const tPlural = t + (t.endsWith('s') ? '' : 's');
  return materials.find((m) => {
    const type = ((m.meta?.type as string) ?? '').toLowerCase();
    const cat = ((m.meta?.category as string) ?? '').toLowerCase();
    return type === t || type === tPlural || cat === t || cat === tPlural || cat.includes(t) || type.includes(t);
  });
}

export default function DetailsDrawer({ nodeData, onClose, onOpenInTab }: DetailsDrawerProps) {
  if (!nodeData) return null;

  const { node, relationships } = nodeData;
  const mats = relationships.materials as MaterialEntry[];

  const nodeType = getNodeType(node.roles);
  const isProduct = node.roles.some((r) => /product/i.test(r));
  const isGoalOrDecision = node.roles.some((r) => /goal|decision/i.test(r));
  const isJob = node.roles.some((r) => /job/i.test(r));
  const isDecision = nodeType === 'Decision';

  const profileMat = isProduct ? matchMaterial(mats, 'Profile') : undefined;
  const rationaleMat = isGoalOrDecision ? matchMaterial(mats, 'Rationale') : undefined;
  const jtbdMat = isJob ? matchMaterial(mats, 'JTBD') : undefined;

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
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem', flexWrap: 'wrap' }}>
            {nodeType && (
              <span
                title={nodeType}
                style={{
                  fontSize: '0.75rem',
                  fontWeight: '600',
                  color: getNodeTypeColor(nodeType),
                  background: `${getNodeTypeColor(nodeType)}18`,
                  padding: '0.25rem 0.5rem',
                  borderRadius: '4px',
                }}
              >
                {nodeType}
              </span>
            )}
            <h3 style={{ margin: 0, fontSize: '1rem', fontWeight: '600' }}>{node.title}</h3>
          </div>
          <div style={{ fontSize: '0.875rem', color: '#666' }}>
            <div>ID: {node.node_id}</div>
            {node.roles.length > 0 && (
              <div style={{ marginTop: '0.25rem' }}>
                Roles: {node.roles.join(', ')}
              </div>
            )}
          </div>
          {isDecision && (
            <div style={{ marginTop: '0.75rem', padding: '0.5rem', background: '#fff8e1', borderRadius: '4px', fontSize: '0.8125rem', color: '#666' }}>
              Decision metadata (date, owner, scope) will appear here when the API supports it.
            </div>
          )}
          {/* Projection: Product → Profile as Description */}
          {profileMat ? (
            <InlineMaterialBlock material={profileMat} nodeId={node.node_id} sectionLabel="Description" />
          ) : node.description ? (
            <div style={{ marginTop: '1rem', padding: '0.75rem', background: '#f5f5f5', borderRadius: '4px' }}>
              <div style={{ fontSize: '0.875rem', fontWeight: '600', marginBottom: '0.5rem' }}>Description</div>
              <div style={{ fontSize: '0.875rem', lineHeight: '1.5' }}>{node.description}</div>
            </div>
          ) : null}
          {/* Projection: Goal/Decision → Rationale as Why */}
          {rationaleMat && (
            <InlineMaterialBlock material={rationaleMat} nodeId={node.node_id} sectionLabel="Why" />
          )}
          {/* Projection: Job → JTBD as Job to be done */}
          {jtbdMat ? (
            <div style={{ marginTop: '1rem', padding: '0.75rem', background: '#e3f2fd', borderRadius: '4px' }}>
              <InlineMaterialBlock material={jtbdMat} nodeId={node.node_id} sectionLabel="Job to be done" />
            </div>
          ) : node.jtbd ? (
            <div style={{ marginTop: '1rem', padding: '0.75rem', background: '#e3f2fd', borderRadius: '4px' }}>
              <div style={{ fontSize: '0.875rem', fontWeight: '600', marginBottom: '0.5rem' }}>Job to be Done</div>
              <div style={{ fontSize: '0.875rem', lineHeight: '1.5' }}>{node.jtbd}</div>
            </div>
          ) : null}
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

        {/* Decisions & Evidence — explicit link types (DECIDED_BY = decision → this node; EVIDENCED_BY = evidence → this node) */}
        {relationships.decisions_and_evidence.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Decisions & Evidence ({relationships.decisions_and_evidence.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {relationships.decisions_and_evidence.map((rel, idx) => {
                const isDecidedBy = rel.type === 'DECIDED_BY';
                const isEvidencedBy = rel.type === 'EVIDENCED_BY';
                const linkLabel = isDecidedBy ? 'Decision' : isEvidencedBy ? 'Evidence' : rel.type;
                const linkColor = isDecidedBy ? '#e65100' : isEvidencedBy ? '#6a1b9a' : '#666';
                return (
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
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                      <span
                        style={{
                          fontSize: '0.6875rem',
                          fontWeight: '600',
                          color: linkColor,
                          background: `${linkColor}18`,
                          padding: '0.125rem 0.375rem',
                          borderRadius: '4px',
                        }}
                      >
                        {linkLabel}
                      </span>
                      <span style={{ fontWeight: '500', fontSize: '0.875rem' }}>{rel.title}</span>
                    </div>
                    <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
                      {rel.type} · click to open in tree
                    </div>
                  </div>
                );
              })}
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
