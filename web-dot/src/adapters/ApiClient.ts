import type { ExpandResult, Operation, DiffResult } from '../types/domain.js';

/**
 * API Client - simple fetch wrapper for /api/* endpoints
 * All requests go to same origin (no CORS issues)
 */
class ApiClient {
  private baseUrl = '/api';

  async healthz(): Promise<{ ok: boolean }> {
    const response = await fetch(`${this.baseUrl}/healthz`);
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.statusText}`);
    }
    return response.json();
  }

  async expand(ids: string[], depth: number = 1): Promise<ExpandResult> {
    const idsParam = ids.join(',');
    const response = await fetch(
      `${this.baseUrl}/expand?ids=${encodeURIComponent(idsParam)}&depth=${depth}`
    );
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(`Expand failed: ${error.error || response.statusText}`);
    }
    return response.json();
  }

  async history(target: string, limit: number = 100): Promise<Operation[]> {
    const response = await fetch(
      `${this.baseUrl}/history?target=${encodeURIComponent(target)}&limit=${limit}`
    );
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(`History failed: ${error.error || response.statusText}`);
    }
    return response.json();
  }

  async diff(aSeq: number, bSeq: number, target?: string): Promise<DiffResult> {
    let url = `${this.baseUrl}/diff?a_seq=${aSeq}&b_seq=${bSeq}`;
    if (target) {
      url += `&target=${encodeURIComponent(target)}`;
    }
    const response = await fetch(url);
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(`Diff failed: ${error.error || response.statusText}`);
    }
    return response.json();
  }
}

export const apiClient = new ApiClient();
