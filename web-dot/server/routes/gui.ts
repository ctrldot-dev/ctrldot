import express from 'express';
import { config } from '../config.js';
import { kernelFetchJson } from '../kernelClient.js';
import { getDescription } from '../descriptions.js';

type ExpandResult = {
  nodes: Array<{ node_id: string; title: string; meta?: Record<string, unknown> }>;
  role_assignments: Array<{ node_id: string; role: string }>;
  links: Array<{ from_node_id: string; to_node_id: string; type: string; link_id: string }>;
  materials: Array<{ material_id: string; node_id: string; content_ref: string; media_type: string; meta?: Record<string, unknown> }>;
};

function rolesByNode(roleAssignments: ExpandResult['role_assignments']): Map<string, string[]> {
  const m = new Map<string, string[]>();
  for (const ra of roleAssignments) {
    const existing = m.get(ra.node_id) || [];
    if (!existing.includes(ra.role)) existing.push(ra.role);
    m.set(ra.node_id, existing);
  }
  return m;
}

export const guiRoutes = express.Router();

// v0.2: GUI config (pinned roots)
guiRoutes.get('/config', (req, res) => {
  res.json({
    label: config.pinnedRoots.label,
    namespace_id: config.namespace, // optional; spec says don't show in main UI, but handy for debugging
    pinned_roots: config.pinnedRoots,
  });
});

// v0.2: connection indicator (kernel reachable)
guiRoutes.get('/healthz', async (req, res) => {
  try {
    const data = await kernelFetchJson<{ ok: boolean }>('/v1/healthz', { query: { namespace_id: undefined } });
    res.json(data);
  } catch (e) {
    res.status(503).json({ ok: false });
  }
});

// v0.2: Products tree view model (CONTAINS only)
guiRoutes.get('/products/tree', async (req, res) => {
  const root = (req.query.root as string) || config.pinnedRoots.products.fieldServe;
  const depth = Number(req.query.depth || 10);

  const expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
    query: { ids: root, depth },
  });

  const nodeById = new Map(expand.nodes.map((n) => [n.node_id, n]));
  const roles = rolesByNode(expand.role_assignments);
  const contains = expand.links.filter((l) => l.type === 'CONTAINS');

  const childrenByParent = new Map<string, string[]>();
  for (const l of contains) {
    const arr = childrenByParent.get(l.from_node_id) || [];
    arr.push(l.to_node_id);
    childrenByParent.set(l.from_node_id, arr);
  }

  // Build signal maps for all nodes
  const alignmentTypes = new Set(['SUPPORTS', 'SATISFIES', 'ADVANCES']);
  const coherenceTypes = new Set(['DEPENDS_ON', 'AFFECTS']);
  const decisionTypes = new Set(['DECIDED_BY', 'EVIDENCED_BY']);
  
  const hasAlignment = new Set<string>();
  const hasCoherence = new Set<string>();
  const hasDecision = new Set<string>();
  const hasMaterials = new Set<string>();

  for (const link of expand.links) {
    if (alignmentTypes.has(link.type)) {
      hasAlignment.add(link.from_node_id);
      hasAlignment.add(link.to_node_id);
    }
    if (coherenceTypes.has(link.type)) {
      hasCoherence.add(link.from_node_id);
      hasCoherence.add(link.to_node_id);
    }
    if (decisionTypes.has(link.type)) {
      hasDecision.add(link.from_node_id);
      hasDecision.add(link.to_node_id);
    }
  }

  for (const mat of expand.materials) {
    hasMaterials.add(mat.node_id);
  }

  type TreeNodeVM = {
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
    children: Array<TreeNodeVM>;
  };

  function build(nodeId: string, visited: Set<string> = new Set()): TreeNodeVM | null {
    if (visited.has(nodeId)) return null;
    visited.add(nodeId);
    const n = nodeById.get(nodeId);
    if (!n) return null;
    const kids: Array<TreeNodeVM> = (childrenByParent.get(nodeId) || [])
      .map((cid) => build(cid, visited))
      .filter((x): x is TreeNodeVM => x !== null);
    const desc = getDescription(nodeId);
    return {
      node_id: n.node_id,
      title: n.title,
      roles: roles.get(n.node_id) || [],
      description: desc.description,
      jtbd: desc.jtbd,
      signals: {
        alignment: hasAlignment.has(nodeId),
        coherence: hasCoherence.has(nodeId),
        decision: hasDecision.has(nodeId),
        materials: hasMaterials.has(nodeId),
      },
      children: kids,
    };
  }

  const tree = build(root) || null;
  res.json({
    root,
    tree,
  });
});

// v0.2: Node detail view model for Details drawer
guiRoutes.get('/node/:nodeId', async (req, res) => {
  const nodeId = req.params.nodeId;
  const depth = Number(req.query.depth || 1);

  const expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
    query: { ids: nodeId, depth },
  });

  const nodeById = new Map(expand.nodes.map((n) => [n.node_id, n]));
  const roles = rolesByNode(expand.role_assignments);

  const center = nodeById.get(nodeId) || null;
  if (!center) {
    res.status(404).json({ error: 'node not found' });
    return;
  }

  // De-dupe links (kernel expand can include duplicates when revisiting nodes)
  const uniqueLinks = (() => {
    const seen = new Set<string>();
    const out: ExpandResult['links'] = [];
    for (const l of expand.links) {
      if (seen.has(l.link_id)) continue;
      seen.add(l.link_id);
      out.push(l);
    }
    return out;
  })();

  const outgoing = uniqueLinks.filter((l) => l.from_node_id === nodeId);
  const incoming = uniqueLinks.filter((l) => l.to_node_id === nodeId);

  const children = outgoing
    .filter((l) => l.type === 'CONTAINS')
    .map((l) => nodeById.get(l.to_node_id))
    .filter(Boolean)
    .map((n) => ({ node_id: n!.node_id, title: n!.title, roles: roles.get(n!.node_id) || [] }))
    .filter((item, idx, arr) => arr.findIndex((x) => x.node_id === item.node_id) === idx);

  const alignmentTypes = new Set(['SUPPORTS', 'SATISFIES', 'ADVANCES']);
  const coherenceTypes = new Set(['DEPENDS_ON', 'AFFECTS']);
  const decisionTypes = new Set(['DECIDED_BY', 'EVIDENCED_BY']);

  function relatedFromLinks(links: typeof outgoing | typeof incoming) {
    return links
      .map((l) => {
        const otherId = l.from_node_id === nodeId ? l.to_node_id : l.from_node_id;
        const other = nodeById.get(otherId);
        if (!other) return null;
        return { type: l.type, node_id: other.node_id, title: other.title, roles: roles.get(other.node_id) || [] };
      })
      .filter(Boolean);
  }

  const relatedAll = [...relatedFromLinks(outgoing), ...relatedFromLinks(incoming)] as Array<{
    type: string;
    node_id: string;
    title: string;
    roles: string[];
  }>;
  const relatedDeduped = relatedAll.filter(
    (item, idx, arr) =>
      arr.findIndex((x) => x.type === item.type && x.node_id === item.node_id) === idx
  );

  const alignment = relatedDeduped.filter((r) => alignmentTypes.has(r.type));
  const coherence = relatedDeduped.filter((r) => coherenceTypes.has(r.type));
  const decisionsAndEvidence = relatedDeduped.filter((r) => decisionTypes.has(r.type));

  const materials = expand.materials
    .filter((m) => m.node_id === nodeId)
    .map((m) => ({ material_id: m.material_id, content_ref: m.content_ref, media_type: m.media_type, meta: m.meta || {} }));

  const desc = getDescription(nodeId);

  res.json({
    node: {
      node_id: center.node_id,
      title: center.title,
      roles: roles.get(center.node_id) || [],
      description: desc.description,
      jtbd: desc.jtbd,
    },
    relationships: {
      children,
      alignment,
      coherence,
      decisions_and_evidence: decisionsAndEvidence,
      materials,
    },
  });
});

// v0.2: Intent view model (SO/AO/TT with connections)
guiRoutes.get('/intent', async (req, res) => {
  const intentIds = [
    config.pinnedRoots.intent.so1,
    config.pinnedRoots.intent.ao1,
    config.pinnedRoots.intent.tt1,
  ];

  const expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
    query: { ids: intentIds.join(','), depth: 2 },
  });

  const nodeById = new Map(expand.nodes.map((n) => [n.node_id, n]));
  const roles = rolesByNode(expand.role_assignments);

  // De-dupe links
  const uniqueLinks = (() => {
    const seen = new Set<string>();
    const out: ExpandResult['links'] = [];
    for (const l of expand.links) {
      if (seen.has(l.link_id)) continue;
      seen.add(l.link_id);
      out.push(l);
    }
    return out;
  })();

  const intentNodes = intentIds
    .map((id) => {
      const node = nodeById.get(id);
      if (!node) return null;

      // Get connections (links where this intent is involved)
      const connections = uniqueLinks
        .filter((l) => l.from_node_id === id || l.to_node_id === id)
        .map((l) => {
          const otherId = l.from_node_id === id ? l.to_node_id : l.from_node_id;
          const other = nodeById.get(otherId);
          if (!other) return null;
          return {
            type: l.type,
            node_id: other.node_id,
            title: other.title,
            roles: roles.get(other.node_id) || [],
          };
        })
        .filter(Boolean) as Array<{ type: string; node_id: string; title: string; roles: string[] }>;

      // Dedupe connections
      const deduped = connections.filter(
        (item, idx, arr) => arr.findIndex((x) => x.type === item.type && x.node_id === item.node_id) === idx
      );

      return {
        node_id: node.node_id,
        title: node.title,
        roles: roles.get(node.node_id) || [],
        connections: deduped,
      };
    })
    .filter(Boolean);

  res.json({ intents: intentNodes });
});

// v0.2: Ledger timeline view model
guiRoutes.get('/ledger', async (req, res) => {
  const target = (req.query.target as string) || config.namespace;
  const limit = Number(req.query.limit || 100);

  const operations = await kernelFetchJson<
    Array<{
      id: string;
      seq: number;
      occurred_at: string;
      actor_id: string;
      changes: Array<{ kind: string; payload?: Record<string, unknown> }>;
    }>
  >('/v1/history', {
    query: { target, limit },
  });

  // Transform to timeline-friendly format
  const timeline = operations.map((op) => {
    const summary =
      op.changes.length > 0
        ? `${op.changes[0].kind}${op.changes.length > 1 ? ` (+${op.changes.length - 1} more)` : ''}`
        : 'No changes';

    // Extract affected node IDs from changes
    const affectedNodes = new Set<string>();
    for (const ch of op.changes) {
      if (ch.payload?.node_id) affectedNodes.add(ch.payload.node_id as string);
      if (ch.payload?.from_node_id) affectedNodes.add(ch.payload.from_node_id as string);
      if (ch.payload?.to_node_id) affectedNodes.add(ch.payload.to_node_id as string);
    }

    return {
      id: op.id,
      seq: op.seq,
      occurred_at: op.occurred_at,
      actor_id: op.actor_id,
      summary,
      affected_nodes: Array.from(affectedNodes),
      changes: op.changes, // Include full changes for details drawer
    };
  });

  res.json({ timeline });
});

// v0.2: Materials index (filesystem-like view)
guiRoutes.get('/materials', async (req, res) => {
  // For now, get all materials by expanding all nodes in namespace
  // This is inefficient but works for demo. In production, we'd want a dedicated materials query.
  const allProductNodes = [config.pinnedRoots.products.fieldServe, config.pinnedRoots.products.assetLink];
  
  const expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
    query: { ids: allProductNodes.join(','), depth: 10 },
  });

  // Group materials by category (from meta or media_type)
  const materials = expand.materials.map((m) => ({
    material_id: m.material_id,
    node_id: m.node_id,
    title: (m.meta?.title as string) || m.content_ref.split('/').pop() || m.material_id,
    content_ref: m.content_ref,
    media_type: m.media_type,
    category: (m.meta?.category as string) || 'Notes', // Default category
    meta: m.meta || {},
  }));

  // Group by category
  const byCategory = new Map<string, typeof materials>();
  for (const m of materials) {
    const cat = m.category;
    const arr = byCategory.get(cat) || [];
    arr.push(m);
    byCategory.set(cat, arr);
  }

  res.json({
    categories: Array.from(byCategory.entries()).map(([category, items]) => ({
      category,
      items,
    })),
  });
});
