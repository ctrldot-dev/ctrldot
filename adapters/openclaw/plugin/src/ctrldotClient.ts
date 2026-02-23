/**
 * HTTP client for Ctrl Dot API.
 */

interface RegisterAgentRequest {
  agent_id: string;
  display_name?: string;
  default_mode?: string;
}

interface ActionProposal {
  agent_id: string;
  session_id?: string;
  intent: {
    title: string;
  };
  action: {
    type: string;
    target: Record<string, any>;
    inputs: Record<string, any>;
  };
  cost: {
    currency: string;
    estimated_gbp: number;
    estimated_tokens: number;
    model: string;
  };
  context: {
    tool: string;
    tags?: string[];
    meta?: Record<string, any>;
  };
}

interface Recommendation {
  kind?: string;
  title?: string;
  summary?: string;
  next_steps?: string[];
  docs_hint?: string;
  tags?: string[];
}

interface Reason {
  code?: string;
  message?: string;
}

interface DecisionResponse {
  decision: "ALLOW" | "WARN" | "THROTTLE" | "DENY" | "STOP";
  reason: string;
  reasons?: Reason[];
  recommendation?: Recommendation;
  model_policy?: string;
  autobundle_path?: string;
  autobundle_trigger?: string;
}

export class CtrlDotClient {
  private baseUrl: string;
  private authToken?: string;
  private agentId: string;
  private agentName: string;
  private agentRegistered: boolean = false;

  constructor(config: {
    ctrldotUrl?: string;
    agentId?: string;
    agentName?: string;
    authToken?: string;
  }) {
    this.baseUrl = config.ctrldotUrl || "http://127.0.0.1:7777";
    this.agentId = config.agentId || "openclaw-agent";
    this.agentName = config.agentName || "OpenClaw Agent";
    this.authToken = config.authToken;
  }

  /**
   * Register agent (idempotent - ignores "already exists" errors).
   */
  async registerAgent(): Promise<void> {
    if (this.agentRegistered) {
      return;
    }

    try {
      const response = await fetch(`${this.baseUrl}/v1/agents/register`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(this.authToken && { Authorization: `Bearer ${this.authToken}` }),
        },
        body: JSON.stringify({
          agent_id: this.agentId,
          display_name: this.agentName,
        }),
      });

      if (response.ok || response.status === 409) {
        // 409 = already exists, which is fine
        this.agentRegistered = true;
        return;
      }

      const text = await response.text();
      throw new Error(`Failed to register agent: ${response.status} ${text}`);
    } catch (error: any) {
      // Ignore connection errors - agent may already be registered
      if (error.message?.includes("fetch")) {
        console.warn(`Ctrl Dot daemon not available: ${error.message}`);
        return;
      }
      throw error;
    }
  }

  /**
   * Propose an action and get a decision.
   */
  async proposeAction(proposal: ActionProposal): Promise<DecisionResponse> {
    // Ensure agent is registered
    await this.registerAgent();

    const response = await fetch(`${this.baseUrl}/v1/actions/propose`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(this.authToken && { Authorization: `Bearer ${this.authToken}` }),
      },
      body: JSON.stringify(proposal),
    });

    if (!response.ok) {
      const text = await response.text();
      let errorData: any = {};
      try {
        errorData = JSON.parse(text);
      } catch {
        // Ignore parse errors
      }

      const decision = errorData.decision || "DENY";
      const reason = errorData.reason || `HTTP ${response.status}: ${text}`;
      throw new Error(`Ctrl Dot denied: ${decision} - ${reason}`);
    }

    return response.json() as Promise<DecisionResponse>;
  }

  getAgentId(): string {
    return this.agentId;
  }
}
