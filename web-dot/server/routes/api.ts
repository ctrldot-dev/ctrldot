import express from 'express';
import { config } from '../config.js';

const router = express.Router();

/**
 * Proxy middleware for /api/* routes
 * Forwards requests to kernel /v1/* and automatically injects namespace
 */
router.all('*', async (req: express.Request, res: express.Response) => {
  // Convert /api/expand to /v1/expand, etc.
  // req.path is relative to the router mount point, so /api/healthz becomes /healthz
  // req.originalUrl includes the full path, so we can extract from there
  const originalPath = req.originalUrl || req.url;
  const kernelPath = originalPath.replace(/^\/api/, '/v1');
  const kernelUrl = `${config.kernelUrl}${kernelPath}`;
  
  // Build query string
  const url = new URL(kernelUrl);
  
  // Copy all query params from original request first
  Object.entries(req.query).forEach(([key, value]) => {
    if (value !== undefined) {
      if (Array.isArray(value)) {
        value.forEach(v => url.searchParams.append(key, String(v)));
      } else {
        url.searchParams.set(key, String(value));
      }
    }
  });
  
  // Always inject namespace_id (will override if already present, which is fine)
  url.searchParams.set('namespace_id', config.namespace);
  
  // Prepare headers (exclude host and connection headers)
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  
  // Copy relevant headers
  const excludeHeaders = ['host', 'connection', 'content-length'];
  Object.entries(req.headers).forEach(([key, value]) => {
    const lowerKey = key.toLowerCase();
    if (!excludeHeaders.includes(lowerKey) && value) {
      headers[key] = Array.isArray(value) ? value[0] : value;
    }
  });
  
  try {
    const fetchOptions: RequestInit = {
      method: req.method,
      headers,
    };
    
    // Add body for POST/PUT/PATCH
    if (['POST', 'PUT', 'PATCH'].includes(req.method) && req.body) {
      fetchOptions.body = JSON.stringify(req.body);
    }
    
    const response = await fetch(url.toString(), fetchOptions);
    
    // Get response data
    const contentType = response.headers.get('content-type');
    let data: unknown;
    
    if (contentType?.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }
    
    // Forward status and headers
    res.status(response.status);
    
    // Copy response headers (exclude some)
    response.headers.forEach((value, key) => {
      const lowerKey = key.toLowerCase();
      if (!['content-encoding', 'content-length', 'transfer-encoding'].includes(lowerKey)) {
        res.setHeader(key, value);
      }
    });
    
    res.json(data);
  } catch (error) {
    console.error('Proxy error:', error);
    res.status(500).json({ 
      error: 'Failed to proxy request to kernel',
      message: error instanceof Error ? error.message : 'Unknown error'
    });
  }
});

export default router;
