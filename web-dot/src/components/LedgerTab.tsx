import React, { useState, useEffect } from 'react';
import { guiClient } from '../adapters/GuiClient.js';
import LedgerEntryDetails from './LedgerEntryDetails.js';

type LedgerTabProps = {
  contextNamespaceId: string | null;
};

export default function LedgerTab({ contextNamespaceId }: LedgerTabProps) {
  const [timeline, setTimeline] = useState<Array<{
    id: string;
    seq: number;
    occurred_at: string;
    actor_id: string;
    summary: string;
    affected_nodes: string[];
    changes?: Array<{ kind: string; payload?: Record<string, unknown> }>;
  }>>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mode, setMode] = useState<'whole' | 'node'>('whole');
  const [selectedEntry, setSelectedEntry] = useState<typeof timeline[0] | null>(null);

  useEffect(() => {
    if (!contextNamespaceId) {
      setTimeline([]);
      return;
    }
    loadHistory();
  }, [contextNamespaceId, mode]);

  const loadHistory = async () => {
    if (!contextNamespaceId) return;
    setLoading(true);
    setError(null);
    try {
      const target = mode === 'whole' ? contextNamespaceId : contextNamespaceId;
      const result = await guiClient.ledger(target, 100);
      setTimeline(result.timeline);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load history');
      console.error('Error loading history:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const handleOpenNode = (nodeId: string) => {
    window.dispatchEvent(new CustomEvent('openInTab', {
      detail: { tab: 'products', nodeId }
    }));
  };

  return (
    <div style={{ display: 'flex', height: '100%', gap: 0 }}>
      <div style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
      <div style={{ marginBottom: '1rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>Ledger</h2>
        <div>
          <button
            onClick={() => setMode('whole')}
            style={{
              padding: '0.5rem 1rem',
              marginRight: '0.5rem',
              background: mode === 'whole' ? '#0066cc' : '#e0e0e0',
              color: mode === 'whole' ? 'white' : '#333',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Whole Ledger
          </button>
          <button
            onClick={() => setMode('node')}
            style={{
              padding: '0.5rem 1rem',
              background: mode === 'node' ? '#0066cc' : '#e0e0e0',
              color: mode === 'node' ? 'white' : '#333',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Selected Node
          </button>
        </div>
      </div>

      {loading && <div className="loading">Loading ledger...</div>}
      {error && <div className="error">Error: {error}</div>}
      {!loading && !error && timeline.length === 0 && (
        <div className="loading">No operations found</div>
      )}
      {!loading && !error && timeline.length > 0 && (
        <div>
          <div style={{ marginBottom: '1rem', color: '#666' }}>
            Showing {timeline.length} operation{timeline.length !== 1 ? 's' : ''}
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {timeline.map((entry) => (
              <div
                key={entry.id}
                onClick={() => setSelectedEntry(entry)}
                style={{
                  padding: '1rem',
                  border: '1px solid #e0e0e0',
                  borderRadius: '4px',
                  background: selectedEntry?.id === entry.id ? '#e3f2fd' : 'white',
                  cursor: 'pointer',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                  <div>
                    <span style={{ fontWeight: '500' }}>Seq {entry.seq}</span>
                    <span style={{ marginLeft: '1rem', color: '#666' }}>
                      {formatDate(entry.occurred_at)}
                    </span>
                  </div>
                  <div style={{ fontSize: '0.875rem', color: '#666' }}>
                    Actor: {entry.actor_id}
                  </div>
                </div>
                <div style={{ color: '#666' }}>{entry.summary}</div>
                {entry.affected_nodes.length > 0 && (
                  <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.5rem' }}>
                    {entry.affected_nodes.length} affected node{entry.affected_nodes.length !== 1 ? 's' : ''}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
      </div>
      {selectedEntry && (
        <LedgerEntryDetails
          entry={selectedEntry}
          onClose={() => setSelectedEntry(null)}
          onOpenNode={handleOpenNode}
        />
      )}
    </div>
  );
}
