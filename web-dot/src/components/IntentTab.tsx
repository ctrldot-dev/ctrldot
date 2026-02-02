import React, { useState, useEffect } from 'react';
import { guiClient, type GuiConfig } from '../adapters/GuiClient.js';

type IntentTabProps = {
  contextNamespaceId: string | null;
};

type IntentEntry = {
  node_id: string;
  title: string;
  roles: string[];
  section: string;
  connections: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
};

export default function IntentTab({ contextNamespaceId }: IntentTabProps) {
  const [sections, setSections] = useState<Array<{ key: string; label: string }>>([]);
  const [intents, setIntents] = useState<IntentEntry[]>([]);
  const [selectedIntent, setSelectedIntent] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [config, setConfig] = useState<GuiConfig | null>(null);

  useEffect(() => {
    if (!contextNamespaceId) {
      setSections([]);
      setIntents([]);
      setSelectedIntent(null);
      return;
    }
    loadConfigAndIntents();
  }, [contextNamespaceId]);

  const loadConfigAndIntents = async () => {
    if (!contextNamespaceId) return;
    setLoading(true);
    setError(null);
    try {
      const cfg = await guiClient.config();
      setConfig(cfg);
      const result = await guiClient.intent(contextNamespaceId);
      const sectionList = result.sections ?? [
        { key: 'strategic_objectives', label: 'Strategic Objectives' },
        { key: 'assurance_obligations', label: 'Assurance Obligations' },
        { key: 'transformation_themes', label: 'Transformation Themes' },
      ];
      const intentList = (result.intents ?? []).map((i) => ({
        ...i,
        section: (i as IntentEntry).section ?? inferSection(i.roles),
      }));
      setSections(sectionList);
      setIntents(intentList);
      const first = intentList[0];
      setSelectedIntent(first ? first.node_id : null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load intents');
      console.error('Error loading intents:', err);
    } finally {
      setLoading(false);
    }
  };

  function inferSection(roles: string[]): string {
    if (!roles.length) return 'strategic_objectives';
    if (roles.some((r) => r.includes('StrategicObjective'))) return 'strategic_objectives';
    if (roles.some((r) => r.includes('AssuranceObligation'))) return 'assurance_obligations';
    if (roles.some((r) => r.includes('TransformationTheme'))) return 'transformation_themes';
    if (roles.some((r) => r === 'Fin.Objective')) return 'strategic_objectives';
    if (roles.some((r) => r === 'Fin.Policy')) return 'assurance_obligations';
    return 'strategic_objectives';
  }

  const selectedIntentData = intents.find((i) => i.node_id === selectedIntent);

  const sectionStyle = (selected: boolean) => ({
    display: 'block',
    width: '100%',
    padding: '0.5rem',
    marginBottom: '0.5rem',
    textAlign: 'left' as const,
    background: selected ? '#e3f2fd' : 'transparent',
    border: '1px solid #e0e0e0',
    borderRadius: '4px',
    cursor: 'pointer',
  });

  return (
    <div style={{ display: 'flex', height: '100%', gap: '1rem' }}>
      <div style={{ width: '250px', borderRight: '1px solid #e0e0e0', padding: '1rem' }}>
        {sections.map((sec) => {
          const sectionIntents = intents.filter((i) => i.section === sec.key);
          const seen = new Set<string>();
          const uniqueIntents = sectionIntents.filter((i) => {
            if (seen.has(i.node_id)) return false;
            seen.add(i.node_id);
            return true;
          });
          return (
            <React.Fragment key={sec.key}>
              <h3 style={{ marginBottom: '1rem', marginTop: sec.key === sections[0]?.key ? 0 : '2rem' }}>
                {sec.label}
              </h3>
              {uniqueIntents.length === 0 && (
                <p style={{ fontSize: '0.875rem', color: '#666' }}>None in this namespace</p>
              )}
              {uniqueIntents.map((intent) => (
                <button
                  key={`${sec.key}-${intent.node_id}`}
                  onClick={() => setSelectedIntent(intent.node_id)}
                  style={sectionStyle(selectedIntent === intent.node_id)}
                >
                  {intent.title}
                </button>
              ))}
            </React.Fragment>
          );
        })}
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
                      key={`${conn.node_id}-${idx}`}
                      role="button"
                      tabIndex={0}
                      onClick={() => {
                        window.dispatchEvent(new CustomEvent('openInTab', {
                          detail: { tab: 'products' as const, nodeId: conn.node_id },
                        }));
                      }}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          window.dispatchEvent(new CustomEvent('openInTab', {
                            detail: { tab: 'products' as const, nodeId: conn.node_id },
                          }));
                        }
                      }}
                      style={{
                        padding: '0.5rem',
                        marginBottom: '0.5rem',
                        border: '1px solid #e0e0e0',
                        borderRadius: '4px',
                        cursor: 'pointer',
                        background: '#fafafa',
                      }}
                    >
                      <div style={{ fontWeight: '500', color: '#1565c0', textDecoration: 'underline' }}>
                        {conn.title}
                      </div>
                      <div style={{ fontSize: '0.875rem', color: '#666' }}>
                        {conn.type} · click to open in Products
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
