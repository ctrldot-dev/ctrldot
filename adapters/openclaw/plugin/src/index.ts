/**
 * Ctrl Dot OpenClaw Plugin
 *
 * Registers guarded tools via OpenClaw's api.registerTool() so they appear
 * in the agent's tool list. Replaces risky built-in exec/web_fetch.
 */

import { CtrlDotClient } from "./ctrldotClient";
import { createCtrlDotProposeTool } from "./tools/ctrldot_propose";
import { createCtrlDotExecTool } from "./tools/ctrldot_exec";
import { createCtrlDotWebFetchTool } from "./tools/ctrldot_web_fetch";

type Api = {
  registerTool: (
    tool: {
      name: string;
      description: string;
      parameters: object;
      execute: (id: unknown, params: Record<string, unknown>) => Promise<{ content: Array<{ type: string; text: string }> }>;
    },
    opts?: { optional?: boolean }
  ) => void;
  config?: {
    plugins?: { entries?: Record<string, { config?: Record<string, unknown> }> };
  };
};

function getPluginConfig(api: Api): { ctrldotUrl?: string; agentId?: string; agentName?: string; authToken?: string } {
  const config = api.config?.plugins?.entries?.ctrldot?.config as Record<string, unknown> | undefined;
  return config
    ? {
        ctrldotUrl: config.ctrldotUrl as string | undefined,
        agentId: config.agentId as string | undefined,
        agentName: config.agentName as string | undefined,
        authToken: config.authToken as string | undefined,
      }
    : {};
}

function registerCtrldot(api: Api): void {
  console.log("[ctrldot] Plugin registering...");
  const pluginConfig = getPluginConfig(api);
  const client = new CtrlDotClient(pluginConfig);

  client.registerAgent().catch((err: Error) => {
    console.warn(`[ctrldot] Failed to register agent: ${err.message}`);
  });

  const tools = [
    createCtrlDotProposeTool(client),
    createCtrlDotExecTool(client),
    createCtrlDotWebFetchTool(client),
  ];

  for (const tool of tools) {
    console.log(`[ctrldot] Registering tool: ${tool.name}`);
    api.registerTool(
      {
        name: tool.name,
        description: tool.description,
        parameters: tool.parameters ?? { type: "object", properties: {} },
        async execute(_id: unknown, params: Record<string, unknown>) {
          console.log(`[ctrldot] Tool ${tool.name} executed with params:`, JSON.stringify(params));
          const text = await tool.execute(params);
          return { content: [{ type: "text", text }] };
        },
      },
      { optional: true }
    );
  }
  console.log("[ctrldot] Plugin registration complete");
}

// OpenClaw may call default(api) or .register(api)
const register = registerCtrldot;
(register as any).register = registerCtrldot;
export default register;

// Legacy export for any loader that expects createCtrlDotPlugin
export function createCtrlDotPlugin(config: Record<string, unknown> = {}): { name: string; version: string; register: (api: Api) => void } {
  const client = new CtrlDotClient(config as any);
  client.registerAgent().catch(() => {});
  const tools = [
    createCtrlDotProposeTool(client),
    createCtrlDotExecTool(client),
    createCtrlDotWebFetchTool(client),
  ];
  return {
    name: "ctrldot",
    version: "0.1.0",
    register(api: Api) {
      for (const tool of tools) {
        api.registerTool(
          {
            name: tool.name,
            description: tool.description,
            parameters: tool.parameters ?? { type: "object", properties: {} },
            async execute(_id: unknown, params: Record<string, unknown>) {
              const text = await tool.execute(params);
              return { content: [{ type: "text", text }] };
            },
          },
          { optional: true }
        );
      }
    },
  };
}
