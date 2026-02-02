export type NamespaceOption = {
  id: string;
  label: string;
  root_node_id: string;
  prefix: string;
};

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
    other?: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
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

  async namespaces(): Promise<{ namespaces: NamespaceOption[]; grouped: Record<string, NamespaceOption[]> }> {
    const res = await fetch(`${this.baseUrl}/namespaces`);
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      const msg = (body as { error?: string })?.error || res.statusText;
      throw new Error(`Failed to load namespaces: ${msg}`);
    }
    return res.json();
  }

  async productsTree(root: string, depth: number = 10, namespaceId?: string): Promise<ProductTreeVM> {
    let url = `${this.baseUrl}/products/tree?root=${encodeURIComponent(root)}&depth=${depth}`;
    if (namespaceId) url += `&namespace_id=${encodeURIComponent(namespaceId)}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to load product tree');
    return res.json();
  }

  async node(nodeId: string, namespaceId?: string): Promise<NodeVM> {
    let url = `${this.baseUrl}/node/${encodeURIComponent(nodeId)}`;
    if (namespaceId) url += `?namespace_id=${encodeURIComponent(namespaceId)}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to load node');
    return res.json();
  }

  async intent(namespaceId?: string): Promise<{
    sections: Array<{ key: string; label: string }>;
    intents: Array<{
      node_id: string;
      title: string;
      roles: string[];
      section: string;
      connections: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
    }>;
  }> {
    let url = `${this.baseUrl}/intent`;
    if (namespaceId) url += `?namespace_id=${encodeURIComponent(namespaceId)}`;
    const res = await fetch(url);
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

  async materials(namespaceId?: string, rootId?: string): Promise<{
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
    const params = new URLSearchParams();
    if (namespaceId) params.set('namespace_id', namespaceId);
    if (rootId) params.set('root', rootId);
    const qs = params.toString();
    const url = qs ? `${this.baseUrl}/materials?${qs}` : `${this.baseUrl}/materials`;
    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to load materials');
    return res.json();
  }

  async narratives(): Promise<{ narratives: NarrativeSummary[] }> {
    const url = `${this.baseUrl}/narratives`;
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 10000);
    try {
      const res = await fetch(url, { signal: controller.signal });
      clearTimeout(timeout);
      if (!res.ok) throw new Error(`Narratives: ${res.status} ${res.statusText}`);
      return res.json();
    } catch (e) {
      clearTimeout(timeout);
      if (e instanceof Error) {
        if (e.name === 'AbortError') throw new Error('Narratives request timed out. Is the server running?');
        throw e;
      }
      throw new Error('Failed to load narratives');
    }
  }

  async narrative(id: string): Promise<Narrative> {
    const res = await fetch(`${this.baseUrl}/narratives/${encodeURIComponent(id)}`);
    if (!res.ok) throw new Error('Failed to load narrative');
    return res.json();
  }
}

export type NarrativeSummary = { id: string; title: string; description: string };
export type NarrativeStep = {
  title: string;
  body: string;
  node_id?: string;
  namespace_id?: string;
};
export type Narrative = {
  id: string;
  title: string;
  description: string;
  steps: NarrativeStep[];
};

export const guiClient = new GuiClient();

