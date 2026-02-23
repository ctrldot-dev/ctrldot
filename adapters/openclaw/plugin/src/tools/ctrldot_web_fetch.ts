/**
 * Guarded web_fetch tool: fetches URLs only after Ctrl Dot allows.
 */

import type { Tool } from "../openclaw-types";
import { CtrlDotClient } from "../ctrldotClient";

const MAX_BYTES = 2 * 1024 * 1024; // 2MB
const ALLOWED_CONTENT_TYPES = [
  "text/plain",
  "text/html",
  "application/json",
  "text/markdown",
  "application/xml",
  "text/xml",
];

export function createCtrlDotWebFetchTool(client: CtrlDotClient): Tool {
  return {
    name: "ctrldot_web_fetch",
    description:
      "Fetch a URL via HTTP GET (guarded by Ctrl Dot). Never use the built-in 'web_fetch' tool - always use this instead.",
    parameters: {
      type: "object",
      properties: {
        url: {
          type: "string",
          description: "URL to fetch",
        },
        maxBytes: {
          type: "number",
          description: "Maximum bytes to fetch (default: 2097152 = 2MB)",
        },
      },
      required: ["url"],
    },
    async execute({ url, maxBytes }: any) {
      // Validate URL
      try {
        new URL(url);
      } catch {
        throw new Error(`Invalid URL: ${url}`);
      }

      // Propose action
      const proposal = {
        agent_id: client.getAgentId(),
        intent: {
          title: `Fetch URL: ${url}`,
        },
        action: {
          type: "tool.call.web_fetch",
          target: {
            url,
            max_bytes: maxBytes || MAX_BYTES,
          },
          inputs: {},
        },
        cost: {
          currency: "GBP",
          estimated_gbp: 0,
          estimated_tokens: 0,
          model: "tool-execution",
        },
        context: {
          tool: "web_fetch",
          tags: ["web", "http"],
          meta: {
            url,
          },
        },
      };

      let decision;
      try {
        decision = await client.proposeAction(proposal);
      } catch (error: any) {
        throw new Error(`Ctrl Dot denied web fetch: ${error.message}`);
      }

      if (decision.decision === "DENY" || decision.decision === "STOP") {
        throw new Error(`Ctrl Dot denied: ${decision.reason}`);
      }

      // Fetch URL
      const response = await fetch(url, {
        method: "GET",
        headers: {
          "User-Agent": "CtrlDot-OpenClaw/0.1.0",
        },
        signal: AbortSignal.timeout(30000), // 30s timeout
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const contentType = response.headers.get("content-type") || "";
      const isAllowedType = ALLOWED_CONTENT_TYPES.some((type) =>
        contentType.includes(type)
      );

      if (!isAllowedType) {
        throw new Error(
          `Content type not allowed: ${contentType}. Allowed types: ${ALLOWED_CONTENT_TYPES.join(", ")}`
        );
      }

      // Read with size limit
      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error("No response body");
      }

      const limit = maxBytes || MAX_BYTES;
      let bytesRead = 0;
      const chunks: Uint8Array[] = [];

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        bytesRead += value.length;
        if (bytesRead > limit) {
          chunks.push(value.slice(0, limit - (bytesRead - value.length)));
          break;
        }
        chunks.push(value);
      }

      const body = Buffer.concat(chunks).toString("utf-8");

      // Return minimal response
      return JSON.stringify(
        {
          url,
          status: response.status,
          statusText: response.statusText,
          contentType,
          headers: Object.fromEntries(response.headers.entries()),
          body: body.substring(0, 50000), // Further truncate for response
          truncated: bytesRead >= limit,
        },
        null,
        2
      );
    },
  };
}
