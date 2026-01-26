// TypeScript types matching the kernel domain types

export interface Node {
  node_id: string;
  title: string;
  meta?: Record<string, unknown>;
}

export interface Link {
  link_id: string;
  from_node_id: string;
  to_node_id: string;
  type: string;
  namespace_id?: string;
  meta?: Record<string, unknown>;
}

export interface RoleAssignment {
  role_assignment_id: string;
  node_id: string;
  namespace_id: string;
  role: string;
  meta?: Record<string, unknown>;
}

export interface Material {
  material_id: string;
  node_id: string;
  content_ref: string;
  media_type: string;
  byte_size: number;
  hash?: string;
  meta?: Record<string, unknown>;
}

export interface Change {
  kind: string;
  namespace_id?: string;
  payload: Record<string, unknown>;
}

export interface Operation {
  id: string;
  seq: number;
  occurred_at: string; // ISO 8601 timestamp
  actor_id: string;
  capabilities: string[];
  plan_id: string;
  plan_hash: string;
  class: number;
  changes: Change[];
}

// Query API types
export interface ExpandResult {
  nodes: Node[];
  role_assignments: RoleAssignment[];
  links: Link[];
  materials: Material[];
}

export interface DiffResult {
  changes: Change[];
}
