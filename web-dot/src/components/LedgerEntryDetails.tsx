import React from 'react';

interface LedgerEntryDetailsProps {
  entry: {
    id: string;
    seq: number;
    occurred_at: string;
    actor_id: string;
    summary: string;
    affected_nodes: string[];
    changes?: Array<{ kind: string; payload?: Record<string, unknown> }>;
  };
  onClose: () => void;
  onOpenNode?: (nodeId: string) => void;
}

export default function LedgerEntryDetails({ entry, onClose, onOpenNode }: LedgerEntryDetailsProps) {
  return (
    <div
      style={{
        width: '500px',
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
        <h2 style={{ margin: 0, fontSize: '1.25rem' }}>Ledger Entry</h2>
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
          <div style={{ fontSize: '0.875rem', color: '#666', marginBottom: '0.5rem' }}>
            Sequence {entry.seq}
          </div>
          <div style={{ fontSize: '0.875rem', color: '#666', marginBottom: '0.5rem' }}>
            {new Date(entry.occurred_at).toLocaleString()}
          </div>
          <div style={{ fontSize: '0.875rem', color: '#666', marginBottom: '1rem' }}>
            Actor: {entry.actor_id}
          </div>
          <div style={{ padding: '0.75rem', background: '#f5f5f5', borderRadius: '4px' }}>
            <div style={{ fontSize: '0.875rem', fontWeight: '600', marginBottom: '0.5rem' }}>Summary</div>
            <div style={{ fontSize: '0.875rem' }}>{entry.summary}</div>
          </div>
        </div>

        {entry.changes && entry.changes.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Changes ({entry.changes.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {entry.changes.map((change, idx) => (
                <div
                  key={idx}
                  style={{
                    padding: '0.75rem',
                    border: '1px solid #e0e0e0',
                    borderRadius: '4px',
                    background: 'white',
                  }}
                >
                  <div style={{ fontWeight: '500', fontSize: '0.875rem', marginBottom: '0.25rem' }}>
                    {change.kind}
                  </div>
                  {change.payload && (
                    <details style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.5rem' }}>
                      <summary style={{ cursor: 'pointer', color: '#0066cc' }}>View payload</summary>
                      <pre style={{ marginTop: '0.5rem', padding: '0.5rem', background: '#f5f5f5', borderRadius: '4px', overflow: 'auto', fontSize: '0.75rem' }}>
                        {JSON.stringify(change.payload, null, 2)}
                      </pre>
                    </details>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {entry.affected_nodes.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600', textTransform: 'uppercase', color: '#666' }}>
              Affected Nodes ({entry.affected_nodes.length})
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {entry.affected_nodes.map((nodeId) => (
                <div
                  key={nodeId}
                  onClick={() => onOpenNode?.(nodeId)}
                  style={{
                    padding: '0.5rem',
                    border: '1px solid #e0e0e0',
                    borderRadius: '4px',
                    cursor: onOpenNode ? 'pointer' : 'default',
                    background: onOpenNode ? '#f5f5f5' : 'white',
                  }}
                >
                  <div style={{ fontSize: '0.875rem', fontFamily: 'monospace' }}>{nodeId}</div>
                  {onOpenNode && (
                    <div style={{ fontSize: '0.75rem', color: '#0066cc', marginTop: '0.25rem' }}>
                      Show in tree
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        <div style={{ marginTop: '1.5rem', paddingTop: '1.5rem', borderTop: '1px solid #e0e0e0' }}>
          <details style={{ fontSize: '0.875rem' }}>
            <summary style={{ cursor: 'pointer', color: '#0066cc', fontWeight: '500' }}>View Raw JSON</summary>
            <pre style={{ marginTop: '0.5rem', padding: '0.5rem', background: '#f5f5f5', borderRadius: '4px', overflow: 'auto', fontSize: '0.75rem' }}>
              {JSON.stringify(entry, null, 2)}
            </pre>
          </details>
        </div>
      </div>
    </div>
  );
}
