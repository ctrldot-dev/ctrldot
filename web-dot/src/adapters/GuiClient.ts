export type GuiConfig = {
  label: string;
  namespace_id?: string;
  pinned_roots: {
    label: string;
    products: { fieldServe: string; assetLink: string };
    intent: { so1: string; ao1: string; tt1: string };
  };
};


export type ProductTreeVM = {
  root: string;
  tree: null | {
    node_id: string;
    title: string;
    roles: string[];
    description?: string;
    jtbd?: string;
    signals?: {
      alignment: boolean;
      coherence: boolean;
      decision: boolean;
      materials: boolean;
    };
    children: ProductTreeVM['tree'][];
  };
};

export type NodeVM = {
  node: { node_id: string; title: string; roles: string[]; description?: string; jtbd?: string };
  relationships: {
    children: Array<{ node_id: string; title: string; roles: string[] }>;
    alignment: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
    coherence: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
    decisions_and_evidence: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
    materials: Array<{ material_id: string; content_ref: string; media_type: string; meta: Record<string, unknown> }>;
  };
};

class GuiClient {
  private baseUrl = '/api';

  async healthz(): Promise<{ ok: boolean }> {
    const res = await fetch(`${this.baseUrl}/healthz`);
    return res.json();
  }

  async config(): Promise<GuiConfig> {
    const res = await fetch(`${this.baseUrl}/config`);
    if (!res.ok) throw new Error('Failed to load config');
    return res.json();
  }

  async productsTree(root: string, depth: number = 10): Promise<ProductTreeVM> {
    const res = await fetch(`${this.baseUrl}/products/tree?root=${encodeURIComponent(root)}&depth=${depth}`);
    if (!res.ok) throw new Error('Failed to load product tree');
    return res.json();
  }

  async node(nodeId: string): Promise<NodeVM> {
    const res = await fetch(`${this.baseUrl}/node/${encodeURIComponent(nodeId)}`);
    if (!res.ok) throw new Error('Failed to load node');
    return res.json();
  }

  async intent(): Promise<{
    intents: Array<{
      node_id: string;
      title: string;
      roles: string[];
      connections: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
    }>;
  }> {
    const res = await fetch(`${this.baseUrl}/intent`);
    if (!res.ok) throw new Error('Failed to load intent');
    return res.json();
  }

  async ledger(target?: string, limit: number = 100): Promise<{
    timeline: Array<{
      id: string;
      seq: number;
      occurred_at: string;
      actor_id: string;
      summary: string;
      affected_nodes: string[];
      changes?: Array<{ kind: string; payload?: Record<string, unknown> }>;
    }>;
  }> {
    let url = `${this.baseUrl}/ledger?limit=${limit}`;
    if (target) url += `&target=${encodeURIComponent(target)}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to load ledger');
    return res.json();
  }

  async materials(): Promise<{
    categories: Array<{
      category: string;
      items: Array<{
        material_id: string;
        node_id: string;
        title: string;
        content_ref: string;
        media_type: string;
        category: string;
        meta: Record<string, unknown>;
      }>;
    }>;
  }> {
    const res = await fetch(`${this.baseUrl}/materials`);
    if (!res.ok) throw new Error('Failed to load materials');
    return res.json();
  }
}

export const guiClient = new GuiClient();

