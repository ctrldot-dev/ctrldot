import { config } from './config.js';

type KernelFetchOptions = {
  method?: string;
  query?: Record<string, string | number | undefined>;
  body?: unknown;
  /** When true, do not inject default namespace_id (e.g. for /v1/namespaces) */
  skipNamespaceInject?: boolean;
};

/**
 * Minimal server-side client for the Kernel HTTP API.
 * Injects namespace_id for endpoints that support it unless skipNamespaceInject or query already has namespace_id.
 */
export async function kernelFetchJson<T>(path: string, opts: KernelFetchOptions = {}): Promise<T> {
  const url = new URL(`${config.kernelUrl}${path}`);

  // copy query params
  if (opts.query) {
    for (const [k, v] of Object.entries(opts.query)) {
      if (v === undefined) continue;
      url.searchParams.set(k, String(v));
    }
  }

  // inject namespace unless caller already provided it or asked to skip
  if (!opts.skipNamespaceInject && !url.searchParams.has('namespace_id')) {
    url.searchParams.set('namespace_id', config.namespace);
  }

  const res = await fetch(url.toString(), {
    method: opts.method || 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });

  const contentType = res.headers.get('content-type') || '';
  const payload = contentType.includes('application/json') ? await res.json() : await res.text();

  if (!res.ok) {
    const msg = typeof payload === 'string' ? payload : (payload?.error || payload?.message || JSON.stringify(payload));
    throw new Error(`Kernel ${res.status} ${res.statusText}: ${msg}`);
  }

  return payload as T;
}

