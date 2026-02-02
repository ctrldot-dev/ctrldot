/**
 * Derive a primary node type from roles for UI (Goal, Job, Decision, Evidence).
 * Roles are typically namespaced (e.g. ProductLedger.Goal, *.Decision).
 */
export type NodeTypeLabel = 'Goal' | 'Job' | 'Decision' | 'Evidence';

export function getNodeType(roles: string[]): NodeTypeLabel | null {
  if (!roles?.length) return null;
  const lower = roles.map((r) => r.toLowerCase());
  if (lower.some((r) => r.includes('.goal') || r.endsWith('goal'))) return 'Goal';
  if (lower.some((r) => r.includes('.job') || r.endsWith('job'))) return 'Job';
  if (lower.some((r) => r.includes('.decision') || r.endsWith('decision'))) return 'Decision';
  if (lower.some((r) => r.includes('.evidence') || r.endsWith('evidence'))) return 'Evidence';
  return null;
}

const NODE_TYPE_COLOR: Record<NodeTypeLabel, string> = {
  Goal: '#2e7d32',
  Job: '#1565c0',
  Decision: '#e65100',
  Evidence: '#6a1b9a',
};

export function getNodeTypeColor(type: NodeTypeLabel): string {
  return NODE_TYPE_COLOR[type];
}

export function getNodeTypeShort(type: NodeTypeLabel): string {
  return type[0];
}
