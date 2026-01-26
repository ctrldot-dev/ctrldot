// Demo configuration - pinned roots for ProductLedger:/Kesteron
// These are the nodes that should be available as starting points

export const demoConfig = {
  // Pinned product roots
  products: [
    { id: 'FieldServe', title: 'P-FS FieldServe' },
    { id: 'AssetLink', title: 'P-AL AssetLink' },
  ],
  
  // Pinned intent roots
  intents: {
    strategicObjectives: [
      { id: 'SO-1', title: 'SO-1 Improve reliability and explainability of customer-critical services' },
    ],
    assuranceObligations: [
      { id: 'AO-1', title: 'AO-1 Maintain traceable decision records for externally visible changes' },
    ],
    transformationThemes: [
      { id: 'TT-1', title: 'TT-1 Increase automation without reducing accountability' },
    ],
  },
};

// Helper to find node ID by title prefix (e.g., "P-FS" -> find node with title starting with "P-FS")
export function findNodeIdByTitlePrefix(prefix: string, nodes: Array<{ node_id: string; title: string }>): string | null {
  const node = nodes.find(n => n.title.startsWith(prefix));
  return node?.node_id || null;
}
