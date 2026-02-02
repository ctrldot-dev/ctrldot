/**
 * Seeded incident narratives for the Decision Ledger demo (WI-1.7).
 * Each step can optionally focus a node in the tree (node_id + namespace_id).
 */

import { config } from './config.js';

export type NarrativeStep = {
  title: string;
  body: string;
  /** When set, selecting this step focuses the node in the tree and opens the details drawer. */
  node_id?: string;
  namespace_id?: string;
};

export type Narrative = {
  id: string;
  title: string;
  description: string;
  steps: NarrativeStep[];
};

const productIncidentNarrative: Narrative = {
  id: 'product-incident-walkthrough',
  title: 'Explaining a past product incident',
  description: 'Trace from an outcome back to the decision and evidence that support it.',
  steps: [
    {
      title: 'Introduction',
      body: 'This walkthrough shows how the Decision Ledger lets you explain a past incident: trace from outcomes back to decisions and evidence, without relying on memory or scattered docs.',
    },
    {
      title: 'Start from the outcome',
      body: 'In the left-hand tree, open a Goal or Job that was affected by an incident. Use Kesteron FieldServe or AssetLink. Select the node to set it as context; the Intent, Ledger, and Materials tabs will show content for that node.',
      namespace_id: 'ProductLedger:/Kesteron/FieldServe',
    },
    {
      title: 'Find the decision',
      body: "In the Details drawer (right), under **Decisions & Evidence**, you see which Decision and Evidence nodes are linked to this node. Decisions show *why* something was done; evidence shows *what* supported it. Click a linked node to open it in the tree.",
    },
    {
      title: 'Trace to evidence',
      body: "From the **Ledger** tab you can see operations that affected this node. Use **Show in tree** on an entry to jump to the related node in the left nav. That keeps your context while you move from outcome → decision → evidence.",
    },
    {
      title: 'Conclusion',
      body: 'You can now explain why the decision was made and what evidence supported it. The same pattern works for financial or other ledgers: outcomes, decisions, and evidence are first-class and navigable.',
    },
  ],
};

const narratives: Narrative[] = [productIncidentNarrative];

export function getNarratives(): Narrative[] {
  return narratives;
}

export function getNarrative(id: string): Narrative | undefined {
  return narratives.find((n) => n.id === id);
}
