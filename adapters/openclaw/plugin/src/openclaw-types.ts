/**
 * Minimal type stubs for OpenClaw plugin API when @openclaw/core is not installed.
 * Replace with real @openclaw/core when available.
 */
export interface Tool {
  name: string;
  description: string;
  parameters?: { type: string; properties?: Record<string, unknown>; required?: string[] };
  execute: (args: Record<string, unknown>) => Promise<string>;
}

export interface Plugin {
  name: string;
  version: string;
  tools: Tool[];
}
