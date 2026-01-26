import { apiClient } from './ApiClient.js';
import { GraphStore } from './GraphStore.js';
import type { Node, Link, RoleAssignment } from '../types/domain.js';

/**
 * High-level adapter that hides API complexity from UI
 * Provides semantic methods like "loadProductTree" instead of "expand with depth"
 */
export class Adapter {
  private graphStore = new GraphStore();

  /**
   * Load a product tree starting from a root node
   * Uses CONTAINS links only to build the hierarchy
   */
  async loadProductTree(rootNodeId: string): Promise<{
    root: Node;
    tree: Array<{ node: Node; children: Node[]; level: number }>;
  }> {
    // Expand with high depth to get full tree
    const result = await apiClient.expand([rootNodeId], 10);
    
    // Update graph store
    this.graphStore.addNodes(result.nodes);
    this.graphStore.addRoles(result.role_assignments);
    this.graphStore.addLinks(result.links);
    
    // Build tree structure using CONTAINS links only
    const root = result.nodes.find(n => n.node_id === rootNodeId);
    if (!root) {
      throw new Error(`Root node ${rootNodeId} not found`);
    }
    
    const tree: Array<{ node: Node; children: Node[]; level: number }> = [];
    const visited = new Set<string>();
    
    const buildTree = (nodeId: string, level: number = 0): void => {
      if (visited.has(nodeId)) return;
      visited.add(nodeId);
      
      const node = this.graphStore.getNode(nodeId);
      if (!node) return;
      
      const children = this.graphStore.getChildren(nodeId, 'CONTAINS');
      tree.push({ node, children, level });
      
      // Recursively add children
      for (const child of children) {
        buildTree(child.node_id, level + 1);
      }
    };
    
    buildTree(rootNodeId);
    
    return { root, tree };
  }

  /**
   * Load details for a specific node including all relationships
   * Groups relationships by type (Children, Alignment, Coherence)
   */
  async loadNodeDetails(nodeId: string): Promise<{
    node: Node;
    roles: RoleAssignment[];
    children: Array<{ node: Node; link: Link }>;
    alignment: Array<{ node: Node; link: Link }>;
    coherence: Array<{ node: Node; link: Link }>;
  }> {
    // Expand with depth 1 to get immediate relationships
    const result = await apiClient.expand([nodeId], 1);
    
    // Update graph store
    this.graphStore.addNodes(result.nodes);
    this.graphStore.addRoles(result.role_assignments);
    this.graphStore.addLinks(result.links);
    
    const node = result.nodes.find(n => n.node_id === nodeId);
    if (!node) {
      throw new Error(`Node ${nodeId} not found`);
    }
    
    const roles = this.graphStore.getRoles(nodeId);
    
    // Group relationships
    const alignmentTypes = ['SUPPORTS', 'SATISFIES', 'ADVANCES'];
    const coherenceTypes = ['DEPENDS_ON', 'AFFECTS', 'DECIDED_BY', 'EVIDENCED_BY'];
    
    const related = this.graphStore.getRelatedNodes(nodeId);
    
    const children = related
      .filter(r => r.link.type === 'CONTAINS' && r.direction === 'to')
      .map(r => ({ node: r.node, link: r.link }));
    
    const alignment = related
      .filter(r => alignmentTypes.includes(r.link.type))
      .map(r => ({ node: r.node, link: r.link }));
    
    const coherence = related
      .filter(r => coherenceTypes.includes(r.link.type))
      .map(r => ({ node: r.node, link: r.link }));
    
    return {
      node,
      roles,
      children,
      alignment,
      coherence,
    };
  }

  /**
   * Load intent summary (SO, AO, TT nodes with their connections)
   */
  async loadIntentSummary(intentNodeIds: string[]): Promise<Array<{
    node: Node;
    roles: RoleAssignment[];
    connections: Array<{ node: Node; link: Link; linkType: string }>;
  }>> {
    // Load all intent nodes
    const result = await apiClient.expand(intentNodeIds, 2);
    
    // Update graph store
    this.graphStore.addNodes(result.nodes);
    this.graphStore.addRoles(result.role_assignments);
    this.graphStore.addLinks(result.links);
    
    return intentNodeIds.map(intentId => {
      const node = this.graphStore.getNode(intentId);
      if (!node) {
        throw new Error(`Intent node ${intentId} not found`);
      }
      
      const roles = this.graphStore.getRoles(intentId);
      const related = this.graphStore.getRelatedNodes(intentId);
      
      const connections = related.map(r => ({
        node: r.node,
        link: r.link,
        linkType: r.link.type,
      }));
      
      return { node, roles, connections };
    });
  }

  /**
   * Load history for a target (node or namespace)
   */
  async loadHistory(target?: string, limit: number = 100): Promise<import('../types/domain.js').Operation[]> {
    const defaultTarget = target || 'ProductLedger:/Kesteron';
    return apiClient.history(defaultTarget, limit);
  }

  /**
   * Load diff between two sequence numbers
   */
  async loadDiff(aSeq: number, bSeq: number, target?: string): Promise<import('../types/domain.js').DiffResult> {
    return apiClient.diff(aSeq, bSeq, target);
  }

  /**
   * Get node from cache (if available)
   */
  getCachedNode(nodeId: string): Node | undefined {
    return this.graphStore.getNode(nodeId);
  }

  /**
   * Find node ID by title prefix
   * Uses history to discover nodes, then expands them to find matches
   */
  async findNodeByTitlePrefix(prefix: string): Promise<string | null> {
    try {
      // Get recent history to discover node IDs
      const operations = await this.loadHistory(undefined, 200);
      
      // Extract node IDs from CreateNode operations
      const nodeIds = new Set<string>();
      for (const op of operations) {
        for (const change of op.changes) {
          if (change.kind === 'CreateNode' && change.payload?.node_id) {
            nodeIds.add(change.payload.node_id as string);
          }
        }
      }
      
      if (nodeIds.size === 0) {
        return null;
      }
      
      // Expand all discovered nodes to get their titles
      const nodeIdArray = Array.from(nodeIds);
      const result = await apiClient.expand(nodeIdArray, 0); // depth 0, just get the nodes
      
      // Find node with matching title prefix
      const matchingNode = result.nodes.find(n => n.title.startsWith(prefix));
      return matchingNode?.node_id || null;
    } catch (err) {
      console.error('Error finding node by prefix:', err);
      return null;
    }
  }
}

export const adapter = new Adapter();
