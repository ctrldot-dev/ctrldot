import type { Node, Link, RoleAssignment } from '../types/domain.js';

/**
 * In-memory graph cache for nodes, links, and role assignments
 * Provides efficient lookups and relationship queries
 */
export class GraphStore {
  private nodesById = new Map<string, Node>();
  private rolesByNodeId = new Map<string, RoleAssignment[]>();
  private linksById = new Map<string, Link>();
  private linksByFromNode = new Map<string, Link[]>();
  private linksByToNode = new Map<string, Link[]>();

  /**
   * Add nodes to the store (deduplicates by node_id)
   */
  addNodes(nodes: Node[]): void {
    for (const node of nodes) {
      this.nodesById.set(node.node_id, node);
    }
  }

  /**
   * Add role assignments to the store
   */
  addRoles(roles: RoleAssignment[]): void {
    for (const role of roles) {
      const existing = this.rolesByNodeId.get(role.node_id) || [];
      // Deduplicate by role_assignment_id
      if (!existing.find(r => r.role_assignment_id === role.role_assignment_id)) {
        existing.push(role);
        this.rolesByNodeId.set(role.node_id, existing);
      }
    }
  }

  /**
   * Add links to the store (deduplicates by link_id)
   */
  addLinks(links: Link[]): void {
    for (const link of links) {
      // Deduplicate by link_id
      if (!this.linksById.has(link.link_id)) {
        this.linksById.set(link.link_id, link);

        // Index by from_node_id
        const fromLinks = this.linksByFromNode.get(link.from_node_id) || [];
        fromLinks.push(link);
        this.linksByFromNode.set(link.from_node_id, fromLinks);

        // Index by to_node_id
        const toLinks = this.linksByToNode.get(link.to_node_id) || [];
        toLinks.push(link);
        this.linksByToNode.set(link.to_node_id, toLinks);
      }
    }
  }

  /**
   * Get a node by ID
   */
  getNode(nodeId: string): Node | undefined {
    return this.nodesById.get(nodeId);
  }

  /**
   * Get all nodes
   */
  getAllNodes(): Node[] {
    return Array.from(this.nodesById.values());
  }

  /**
   * Get children of a node (outgoing CONTAINS links)
   */
  getChildren(nodeId: string, linkType?: string): Node[] {
    const links = this.linksByFromNode.get(nodeId) || [];
    const filtered = linkType 
      ? links.filter(l => l.type === linkType)
      : links;
    
    return filtered
      .map(link => this.nodesById.get(link.to_node_id))
      .filter((node): node is Node => node !== undefined);
  }

  /**
   * Get parents of a node (incoming CONTAINS links)
   */
  getParents(nodeId: string, linkType?: string): Node[] {
    const links = this.linksByToNode.get(nodeId) || [];
    const filtered = linkType
      ? links.filter(l => l.type === linkType)
      : links;
    
    return filtered
      .map(link => this.nodesById.get(link.from_node_id))
      .filter((node): node is Node => node !== undefined);
  }

  /**
   * Get related nodes (all links, optionally filtered by type)
   */
  getRelatedNodes(nodeId: string, linkType?: string): Array<{ node: Node; link: Link; direction: 'from' | 'to' }> {
    const fromLinks = this.linksByFromNode.get(nodeId) || [];
    const toLinks = this.linksByToNode.get(nodeId) || [];
    
    const allLinks = [...fromLinks, ...toLinks];
    const filtered = linkType
      ? allLinks.filter(l => l.type === linkType)
      : allLinks;
    
    return filtered.map(link => {
      const relatedNodeId = link.from_node_id === nodeId 
        ? link.to_node_id 
        : link.from_node_id;
      const node = this.nodesById.get(relatedNodeId);
      if (!node) return null;
      
      return {
        node,
        link,
        direction: link.from_node_id === nodeId ? 'to' : 'from',
      };
    }).filter((item): item is { node: Node; link: Link; direction: 'from' | 'to' } => item !== null);
  }

  /**
   * Get role assignments for a node
   */
  getRoles(nodeId: string): RoleAssignment[] {
    return this.rolesByNodeId.get(nodeId) || [];
  }

  /**
   * Clear all data
   */
  clear(): void {
    this.nodesById.clear();
    this.rolesByNodeId.clear();
    this.linksById.clear();
    this.linksByFromNode.clear();
    this.linksByToNode.clear();
  }
}
