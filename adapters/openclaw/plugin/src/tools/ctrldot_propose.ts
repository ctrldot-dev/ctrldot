/**
 * Diagnostic tool: propose an action and return decision JSON.
 */

import type { Tool } from "../openclaw-types";
import { CtrlDotClient } from "../ctrldotClient";

export function createCtrlDotProposeTool(client: CtrlDotClient): Tool {
  return {
    name: "ctrldot_propose",
    description: "Propose an action to Ctrl Dot and get a decision. Use this to check if an action would be allowed before executing it.",
    parameters: {
      type: "object",
      properties: {
        kind: {
          type: "string",
          description: "Action kind (e.g., 'tool.call.exec', 'tool.call.web_fetch')",
        },
        name: {
          type: "string",
          description: "Tool/action name",
        },
        args: {
          type: "object",
          description: "Action arguments",
        },
        meta: {
          type: "object",
          description: "Optional metadata",
        },
      },
      required: ["kind", "name", "args"],
    },
    async execute({ kind, name, args, meta }: any) {
      const proposal = {
        agent_id: client.getAgentId(),
        intent: {
          title: `Propose: ${name}`,
        },
        action: {
          type: kind,
          target: args,
          inputs: {},
        },
        cost: {
          currency: "GBP",
          estimated_gbp: 0,
          estimated_tokens: 0,
          model: "tool-execution",
        },
        context: {
          tool: name,
          tags: [kind, name],
          meta: meta || {},
        },
      };

      try {
        const decision = await client.proposeAction(proposal);
        let out = JSON.stringify(decision, null, 2);
        if (decision.recommendation && (decision.decision === "DENY" || decision.decision === "STOP")) {
          const rec = decision.recommendation;
          const summary = rec.summary || rec.title || "";
          const steps = (rec.next_steps || []).filter(Boolean);
          const parts: string[] = [];
          if (summary) parts.push(`Recommendation: ${summary}`);
          if (steps.length) parts.push(`Next steps:\n${steps.map((s) => `  - ${s}`).join("\n")}`);
          if (decision.autobundle_path) parts.push(`Bundle: ${decision.autobundle_path}`);
          if (parts.length) out = parts.join("\n\n") + "\n\n---\n\n" + out;
        }
        return out;
      } catch (error: any) {
        return JSON.stringify(
          {
            decision: "DENY",
            reason: error.message,
            error: true,
          },
          null,
          2
        );
      }
    },
  };
}
