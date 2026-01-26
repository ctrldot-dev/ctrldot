// Simple in-memory description store
// In production, this could be replaced with a database or derived from Materials

type DescriptionStore = {
  [nodeId: string]: {
    description?: string;
    jtbd?: string; // Jobs to be Done (for Job nodes)
  };
};

const descriptions: DescriptionStore = {
  // Example descriptions - in production, these would be loaded from a file or database
  // Format: node_id (without 'node:' prefix) -> { description, jtbd? }
};

export function getDescription(nodeId: string): { description?: string; jtbd?: string } {
  // Normalize node ID (strip 'node:' prefix if present)
  const normalizedId = nodeId.replace(/^node:/, '');
  return descriptions[normalizedId] || {};
}

export function setDescription(nodeId: string, description: string, jtbd?: string): void {
  const normalizedId = nodeId.replace(/^node:/, '');
  descriptions[normalizedId] = { description, jtbd };
}

export function getAllDescriptions(): DescriptionStore {
  return { ...descriptions };
}
