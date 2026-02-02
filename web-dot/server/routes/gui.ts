import express from 'express';
import { config } from '../config.js';
import { kernelFetchJson } from '../kernelClient.js';
import { getDescription } from '../descriptions.js';
import { getNarratives, getNarrative } from '../narratives.js';

type NamespaceOption = {
  id: string;
  label: string;
  root_node_id: string;
  prefix: string;
};

type RoleAssignmentVM = { node_id: string; role: string; meta?: Record<string, unknown> };

type ExpandResult = {
  nodes: Array<{ node_id: string; title: string; meta?: Record<string, unknown> }>;
  role_assignments: RoleAssignmentVM[];
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

/** Role assignments with meta (for Fin.Policy policy_kind). */
function roleAssignmentsWithMeta(roleAssignments: ExpandResult['role_assignments']): Map<string, RoleAssignmentVM[]> {
  const m = new Map<string, RoleAssignmentVM[]>();
  for (const ra of roleAssignments) {
    const arr = m.get(ra.node_id) || [];
    arr.push(ra);
    m.set(ra.node_id, arr);
  }
  return m;
}

function isProductLedgerNamespace(namespaceId: string): boolean {
  return namespaceId.startsWith('ProductLedger:');
}

function isFinLedgerNamespace(namespaceId: string): boolean {
  return namespaceId.startsWith('FinLedger:');
}

/** Product intent roles (and legacy EnterpriseIntent.*). */
const PRODUCT_SO_ROLES = ['Product.StrategicObjective', 'EnterpriseIntent.StrategicObjective'];
const PRODUCT_AO_ROLES = ['Product.AssuranceObligation', 'EnterpriseIntent.AssuranceObligation'];
const PRODUCT_TT_ROLES = ['Product.TransformationTheme', 'EnterpriseIntent.TransformationTheme'];

function hasRole(nodeRoles: string[], roleSet: string[]): boolean {
  return nodeRoles.some((r) => roleSet.some((allowed) => r === allowed || r.endsWith(allowed)));
}

/** Section key for grouping intents. */
type IntentSectionKey = 'strategic_objectives' | 'assurance_obligations' | 'transformation_themes' | 'policies';

export const guiRoutes = express.Router();

// v0.2: GUI config (pinned roots + available namespaces)
guiRoutes.get('/config', (req, res) => {
  const nsEntry = config.availableNamespaces.find((n) => n.id === config.namespace);
  const label = nsEntry?.label ?? config.pinnedRoots.label;
  res.json({
    label,
    namespace_id: config.namespace,
    pinned_roots: config.pinnedRoots,
    available_namespaces: config.availableNamespaces,
  });
});

// Ledger / Namespace browser: list namespaces from kernel only (WI-4.4: load only explicitly defined demo content)
guiRoutes.get('/namespaces', async (req, res) => {
  try {
    const list = await kernelFetchJson<{ namespace_ids: string[] }>('/v1/namespaces', {
      query: {},
      skipNamespaceInject: true,
    });
    const namespaces: NamespaceOption[] = [];
    for (const id of list.namespace_ids || []) {
      try {
        const root = await kernelFetchJson<{ node_id: string; title: string; role: string }>('/v1/namespace_root', {
          query: { namespace_id: id },
          skipNamespaceInject: true,
        });
        const prefix = id.includes(':') ? id.split(':')[0]! : id;
        namespaces.push({
          id,
          label: root.title || id,
          root_node_id: root.node_id,
          prefix,
        });
      } catch {
        // No root (e.g. empty namespace) — still list with id as label
        const prefix = id.includes(':') ? id.split(':')[0]! : id;
        namespaces.push({ id, label: id, root_node_id: '', prefix });
      }
    }
    namespaces.sort((a, b) => a.id.localeCompare(b.id));
    // Group by prefix for UI
    const grouped = namespaces.reduce<Record<string, NamespaceOption[]>>((acc, ns) => {
      const p = ns.prefix || 'Other';
      if (!acc[p]) acc[p] = [];
      acc[p].push(ns);
      return acc;
    }, {});
    res.json({ namespaces, grouped: grouped });
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e);
    const hint = msg.includes('fetch') || msg.includes('ECONNREFUSED')
      ? ' (Is the kernel running on the configured KERNEL_URL?)'
      : '';
    res.status(503).json({ error: msg + hint });
  }
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

// v0.2: Incident narratives (WI-1.7)
guiRoutes.get('/narratives', (req, res) => {
  try {
    const list = getNarratives();
    res.json({ narratives: list });
  } catch (e) {
    console.error('GET /api/narratives:', e);
    res.status(500).json({ error: 'Failed to load narratives', detail: String(e) });
  }
});
guiRoutes.get('/narratives/:id', (req, res) => {
  try {
    const n = getNarrative(req.params.id);
    if (!n) return res.status(404).json({ error: 'Narrative not found' });
    res.json(n);
  } catch (e) {
    console.error('GET /api/narratives/:id:', e);
    res.status(500).json({ error: 'Failed to load narrative', detail: String(e) });
  }
});

// v0.2: Products tree view model (CONTAINS only). Supports namespace_id for tabbed ledgers.
guiRoutes.get('/products/tree', async (req, res) => {
  const root = (req.query.root as string) || config.pinnedRoots.products.fieldServe;
  const depth = Number(req.query.depth || 10);
  const namespaceId = (req.query.namespace_id as string) || config.namespace;

  if (!root?.trim()) {
    res.status(400).json({ error: 'root is required (use namespace list from /api/namespaces)' });
    return;
  }

  const expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
    query: { ids: root, depth, namespace_id: namespaceId },
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

// v0.2: Node detail view model for Details drawer. Supports namespace_id for tabbed ledgers.
guiRoutes.get('/node/:nodeId', async (req, res) => {
  const nodeId = req.params.nodeId;
  const depth = Number(req.query.depth || 1);
  const namespaceId = (req.query.namespace_id as string) || config.namespace;

  const expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
    query: { ids: nodeId, depth, namespace_id: namespaceId },
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
  const childIds = new Set(children.map((c) => c.node_id));
  // All other link types: CONTAINS incoming (parent), and any type not in alignment/coherence/decisions (exclude CONTAINS outgoing, already in children)
  const other = relatedDeduped.filter(
    (r) =>
      !alignmentTypes.has(r.type) &&
      !coherenceTypes.has(r.type) &&
      !decisionTypes.has(r.type) &&
      !(r.type === 'CONTAINS' && childIds.has(r.node_id))
  );

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
      other,
      materials,
    },
  });
});

// v0.2: Intent view model by ledger type. Discovers intent nodes by querying nodes by role within the selected namespace/root.
// Product ledger: Product.StrategicObjective, Product.AssuranceObligation, Product.TransformationTheme (and EnterpriseIntent.*).
// Fin ledger: Fin.Objective (SO), Fin.Policy (AO), Policies (Fin.Policy with meta.policy_kind == "operational_controls" or all Fin.Policy).
guiRoutes.get('/intent', async (req, res) => {
  const namespaceId = (req.query.namespace_id as string) || config.namespace;

  let expand: ExpandResult;
  try {
    const rootResp = await kernelFetchJson<{ node_id: string; title: string; role: string }>('/v1/namespace_root', {
      query: { namespace_id: namespaceId },
      skipNamespaceInject: true,
    });
    const rootId = rootResp?.node_id;
    if (!rootId) {
      res.json({ sections: [], intents: [] });
      return;
    }
    expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
      query: { ids: rootId, depth: 5, namespace_id: namespaceId },
    });
  } catch {
    res.json({ sections: [], intents: [] });
    return;
  }

  const nodeById = new Map(expand.nodes.map((n) => [n.node_id, n]));
  const roles = rolesByNode(expand.role_assignments);
  const rolesWithMeta = roleAssignmentsWithMeta(expand.role_assignments);

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

  type IntentNode = {
    node_id: string;
    title: string;
    roles: string[];
    section: IntentSectionKey;
    connections: Array<{ type: string; node_id: string; title: string; roles: string[] }>;
  };

  const intentNodesBySection = new Map<IntentSectionKey, IntentNode[]>();
  const sections: Array<{ key: IntentSectionKey; label: string }> = [];

  if (isProductLedgerNamespace(namespaceId)) {
    sections.push(
      { key: 'strategic_objectives', label: 'Strategic Objectives' },
      { key: 'assurance_obligations', label: 'Assurance Obligations' },
      { key: 'transformation_themes', label: 'Transformation Themes' }
    );
    for (const n of expand.nodes) {
      const nodeRoles = roles.get(n.node_id) || [];
      let section: IntentSectionKey | null = null;
      if (hasRole(nodeRoles, PRODUCT_SO_ROLES)) section = 'strategic_objectives';
      else if (hasRole(nodeRoles, PRODUCT_AO_ROLES)) section = 'assurance_obligations';
      else if (hasRole(nodeRoles, PRODUCT_TT_ROLES)) section = 'transformation_themes';
      if (!section) continue;

      const connections = uniqueLinks
        .filter((l) => l.from_node_id === n.node_id || l.to_node_id === n.node_id)
        .map((l) => {
          const otherId = l.from_node_id === n.node_id ? l.to_node_id : l.from_node_id;
          const other = nodeById.get(otherId);
          if (!other) return null;
          return {
            type: l.type,
            node_id: other.node_id,
            title: other.title,
            roles: roles.get(other.node_id) || [],
          };
        })
        .filter(Boolean) as IntentNode['connections'];
      const deduped = connections.filter(
        (item, idx, arr) => arr.findIndex((x) => x.type === item.type && x.node_id === item.node_id) === idx
      );

      const intent: IntentNode = {
        node_id: n.node_id,
        title: n.title,
        roles: nodeRoles,
        section,
        connections: deduped,
      };
      const arr = intentNodesBySection.get(section) || [];
      arr.push(intent);
      intentNodesBySection.set(section, arr);
    }
  } else if (isFinLedgerNamespace(namespaceId)) {
    sections.push(
      { key: 'strategic_objectives', label: 'Strategic Objectives' },
      { key: 'assurance_obligations', label: 'Assurance Obligations' },
      { key: 'policies', label: 'Policies' }
    );
    for (const n of expand.nodes) {
      const nodeRoles = roles.get(n.node_id) || [];
      const sectionsForNode: IntentSectionKey[] = [];
      if (nodeRoles.includes('Fin.Objective')) sectionsForNode.push('strategic_objectives');
      if (nodeRoles.includes('Fin.Policy')) {
        sectionsForNode.push('assurance_obligations');
        sectionsForNode.push('policies');
      }
      if (sectionsForNode.length === 0) continue;

      const connections = uniqueLinks
        .filter((l) => l.from_node_id === n.node_id || l.to_node_id === n.node_id)
        .map((l) => {
          const otherId = l.from_node_id === n.node_id ? l.to_node_id : l.from_node_id;
          const other = nodeById.get(otherId);
          if (!other) return null;
          return {
            type: l.type,
            node_id: other.node_id,
            title: other.title,
            roles: roles.get(other.node_id) || [],
          };
        })
        .filter(Boolean) as IntentNode['connections'];
      const deduped = connections.filter(
        (item, idx, arr) => arr.findIndex((x) => x.type === item.type && x.node_id === item.node_id) === idx
      );

      const intent: IntentNode = {
        node_id: n.node_id,
        title: n.title,
        roles: nodeRoles,
        section: sectionsForNode[0]!,
        connections: deduped,
      };
      for (const section of sectionsForNode) {
        const arr = intentNodesBySection.get(section) || [];
        arr.push({ ...intent, section });
        intentNodesBySection.set(section, arr);
      }
    }
  } else {
    res.json({ sections: [], intents: [] });
    return;
  }

  const intents: IntentNode[] = [];
  for (const s of sections) {
    const list = intentNodesBySection.get(s.key) || [];
    for (const i of list) intents.push(i);
  }

  res.json({ sections, intents });
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

// v0.2: Materials index (filesystem-like view). Optional namespace_id and root scope to that subtree.
// When namespace_id is set but root is missing, discovers the namespace root so materials for that namespace are shown.
guiRoutes.get('/materials', async (req, res) => {
  const namespaceId = req.query.namespace_id as string | undefined;
  let rootId = req.query.root as string | undefined;

  // When we have namespace but no root, get the namespace root so we can list materials in that namespace
  if (namespaceId && !rootId) {
    try {
      const rootResp = await kernelFetchJson<{ node_id: string; title: string; role: string }>('/v1/namespace_root', {
        query: { namespace_id: namespaceId },
        skipNamespaceInject: true,
      });
      rootId = rootResp?.node_id || undefined;
    } catch {
      rootId = undefined;
    }
  }

  let allNodeIds: Set<string>;
  if (namespaceId && rootId) {
    // Context-scoped: expand from root in that namespace
    const expandFromRoot = await kernelFetchJson<ExpandResult>('/v1/expand', {
      query: { ids: rootId, depth: 10, namespace_id: namespaceId },
    });
    allNodeIds = new Set(expandFromRoot.nodes.map((n) => n.node_id));
  } else {
    // Default: config-based roots (may be empty after Product Ledger split; nav provides root/namespace)
    const productRoots = [
      config.pinnedRoots.products.fieldServe,
      config.pinnedRoots.products.assetLink,
    ].filter(Boolean);
    const intentRoots = [
      config.pinnedRoots.intent.so1,
      config.pinnedRoots.intent.ao1,
      config.pinnedRoots.intent.tt1,
    ].filter(Boolean);
    const idsToExpand = [...productRoots, ...intentRoots];
    if (idsToExpand.length === 0) {
      allNodeIds = new Set();
    } else {
      const expandProducts = await kernelFetchJson<ExpandResult>('/v1/expand', {
        query: {
          ids: productRoots.length ? productRoots.join(',') : intentRoots.join(','),
          depth: 10,
          namespace_id: config.namespace,
        },
      });
      allNodeIds = new Set(expandProducts.nodes.map((n) => n.node_id));
      intentRoots.forEach((id) => allNodeIds.add(id));
    }
  }

  const allMaterials: ExpandResult['materials'] = [];
  const seenMaterials = new Set<string>();
  const nodeArray = Array.from(allNodeIds);
  const nsForExpand = namespaceId || config.namespace;

  for (let i = 0; i < nodeArray.length; i += 50) {
    const batch = nodeArray.slice(i, i + 50);
    const expand = await kernelFetchJson<ExpandResult>('/v1/expand', {
      query: { ids: batch.join(','), depth: 0, namespace_id: nsForExpand },
    });
    for (const m of expand.materials) {
      if (!seenMaterials.has(m.material_id)) {
        seenMaterials.add(m.material_id);
        allMaterials.push(m);
      }
    }
  }

  // Group materials by category (from meta or media_type)
  const materials = allMaterials.map((m) => ({
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
