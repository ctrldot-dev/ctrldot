import React, { useState, useEffect } from 'react';
import { guiClient, type GuiConfig } from '../adapters/GuiClient.js';

export default function IntentTab() {
  const [intents, setIntents] = useState<Array<{
    node_id: string;
    title: string;
    roles: string[];
    connections: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
  }>>([]);
  const [selectedIntent, setSelectedIntent] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [config, setConfig] = useState<GuiConfig | null>(null);

  useEffect(() => {
    loadConfigAndIntents();
  }, []);

  const loadConfigAndIntents = async () => {
    setLoading(true);
    setError(null);
    try {
      const cfg = await guiClient.config();
      setConfig(cfg);
      
      const result = await guiClient.intent();
      setIntents(result.intents);
      if (result.intents.length > 0) {
        setSelectedIntent(result.intents[0].node_id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load intents');
      console.error('Error loading intents:', err);
    } finally {
      setLoading(false);
    }
  };

  const selectedIntentData = intents.find(i => i.node_id === selectedIntent);
  
  // Group intents by type
  const so1 = intents.find(i => i.node_id === config?.pinned_roots.intent.so1);
  const ao1 = intents.find(i => i.node_id === config?.pinned_roots.intent.ao1);
  const tt1 = intents.find(i => i.node_id === config?.pinned_roots.intent.tt1);

  return (
    <div style={{ display: 'flex', height: '100%', gap: '1rem' }}>
      <div style={{ width: '250px', borderRight: '1px solid #e0e0e0', padding: '1rem' }}>
        <h3 style={{ marginBottom: '1rem' }}>Strategic Objectives</h3>
        {so1 && (
          <button
            onClick={() => setSelectedIntent(so1.node_id)}
            style={{
              display: 'block',
              width: '100%',
              padding: '0.5rem',
              marginBottom: '0.5rem',
              textAlign: 'left',
              background: selectedIntent === so1.node_id ? '#e3f2fd' : 'transparent',
              border: '1px solid #e0e0e0',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            {so1.title}
          </button>
        )}
        
        <h3 style={{ marginTop: '2rem', marginBottom: '1rem' }}>Assurance Obligations</h3>
        {ao1 && (
          <button
            onClick={() => setSelectedIntent(ao1.node_id)}
            style={{
              display: 'block',
              width: '100%',
              padding: '0.5rem',
              marginBottom: '0.5rem',
              textAlign: 'left',
              background: selectedIntent === ao1.node_id ? '#e3f2fd' : 'transparent',
              border: '1px solid #e0e0e0',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            {ao1.title}
          </button>
        )}
        
        <h3 style={{ marginTop: '2rem', marginBottom: '1rem' }}>Transformation Themes</h3>
        {tt1 && (
          <button
            onClick={() => setSelectedIntent(tt1.node_id)}
            style={{
              display: 'block',
              width: '100%',
              padding: '0.5rem',
              marginBottom: '0.5rem',
              textAlign: 'left',
              background: selectedIntent === tt1.node_id ? '#e3f2fd' : 'transparent',
              border: '1px solid #e0e0e0',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            {tt1.title}
          </button>
        )}
      </div>
      <div style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        {loading && <div className="loading">Loading intents...</div>}
        {error && <div className="error">Error: {error}</div>}
        {!loading && !error && selectedIntentData && (
          <div>
            <h2>{selectedIntentData.title}</h2>
            <div style={{ marginTop: '1rem' }}>
              <h3>Connections ({selectedIntentData.connections.length})</h3>
              {selectedIntentData.connections.length === 0 ? (
                <p>No connections found</p>
              ) : (
                <div style={{ marginTop: '0.5rem' }}>
                  {selectedIntentData.connections.map((conn, idx) => (
                    <div
                      key={idx}
                      style={{
                        padding: '0.5rem',
                        marginBottom: '0.5rem',
                        border: '1px solid #e0e0e0',
                        borderRadius: '4px',
                      }}
                    >
                      <div style={{ fontWeight: '500' }}>{conn.title}</div>
                      <div style={{ fontSize: '0.875rem', color: '#666' }}>
                        Link: {conn.type}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
        {!loading && !error && !selectedIntentData && (
          <div className="loading">Select an intent to view details</div>
        )}
      </div>
    </div>
  );
}
