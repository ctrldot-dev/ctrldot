/**
 * Guarded exec tool: runs shell commands only after Ctrl Dot allows.
 */

import { spawn } from "child_process";
import type { Tool } from "../openclaw-types";
import { CtrlDotClient } from "../ctrldotClient";

const DEFAULT_TIMEOUT_MS = 60000;
const MAX_OUTPUT_CHARS = 10000;

export function createCtrlDotExecTool(client: CtrlDotClient): Tool {
  return {
    name: "ctrldot_exec",
    description:
      "Execute a shell command (guarded by Ctrl Dot). Never use the built-in 'exec' tool - always use this instead.",
    parameters: {
      type: "object",
      properties: {
        command: {
          type: "string",
          description: "Shell command to execute",
        },
        cwd: {
          type: "string",
          description: "Working directory (optional)",
        },
        timeoutMs: {
          type: "number",
          description: "Timeout in milliseconds (default: 60000)",
        },
      },
      required: ["command"],
    },
    async execute({ command, cwd, timeoutMs }: any) {
      // Propose action
      const proposal = {
        agent_id: client.getAgentId(),
        intent: {
          title: `Execute command: ${command}`,
        },
        action: {
          type: "tool.call.exec",
          target: {
            command,
            cwd: cwd || process.cwd(),
            timeout_ms: timeoutMs || DEFAULT_TIMEOUT_MS,
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
          tool: "exec",
          tags: ["exec", "shell"],
          meta: {
            cwd: cwd || process.cwd(),
          },
        },
      };

      let decision;
      try {
        decision = await client.proposeAction(proposal);
      } catch (error: any) {
        throw new Error(`Ctrl Dot denied command execution: ${error.message}`);
      }

      if (decision.decision === "DENY" || decision.decision === "STOP") {
        throw new Error(`Ctrl Dot denied: ${decision.reason}`);
      }

      // Execute command
      return new Promise<string>((resolve, reject) => {
        const timeout = timeoutMs || DEFAULT_TIMEOUT_MS;
        const proc = spawn(command, {
          shell: true,
          cwd: cwd || process.cwd(),
        });

        let stdout = "";
        let stderr = "";
        let timeoutHandle: NodeJS.Timeout;

        const cleanup = () => {
          if (timeoutHandle) clearTimeout(timeoutHandle);
          proc.kill();
        };

        timeoutHandle = setTimeout(() => {
          cleanup();
          reject(new Error(`Command timed out after ${timeout}ms`));
        }, timeout);

        proc.stdout.on("data", (data) => {
          stdout += data.toString();
          if (stdout.length > MAX_OUTPUT_CHARS) {
            stdout = stdout.substring(0, MAX_OUTPUT_CHARS) + "... (truncated)";
            cleanup();
          }
        });

        proc.stderr.on("data", (data) => {
          stderr += data.toString();
          if (stderr.length > MAX_OUTPUT_CHARS) {
            stderr = stderr.substring(0, MAX_OUTPUT_CHARS) + "... (truncated)";
          }
        });

        proc.on("close", (code) => {
          cleanup();
          const output = `Exit code: ${code}\n\nSTDOUT:\n${stdout}\n\nSTDERR:\n${stderr}`;
          resolve(output);
        });

        proc.on("error", (error) => {
          cleanup();
          reject(error);
        });
      });
    },
  };
}
