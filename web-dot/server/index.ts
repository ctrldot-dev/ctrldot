import express from 'express';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { config } from './config.js';
import { guiRoutes } from './routes/gui.js';
import { kernelProxy } from './routes/kernelProxy.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const app = express();

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// v0.2 GUI API routes (must come before static files)
app.use('/api', guiRoutes);

// Legacy kernel proxy (debug only)
app.use('/kernel', kernelProxy);

// Serve static files from public directory
const publicPath = join(__dirname, '../public');
app.use(express.static(publicPath));

// Serve compiled JavaScript from dist/src directory
const distSrcPath = join(__dirname, '../dist/src');
app.use('/src', express.static(distSrcPath, {
  setHeaders: (res, path) => {
    if (path.endsWith('.js')) {
      res.setHeader('Content-Type', 'application/javascript');
    }
  }
}));

// Serve materials directory (for material content files)
const materialsPath = join(__dirname, '../../materials');
app.use('/materials', express.static(materialsPath, {
  setHeaders: (res, path) => {
    if (path.endsWith('.md')) {
      res.setHeader('Content-Type', 'text/markdown; charset=utf-8');
    }
  }
}));

// Fallback: serve index.html for SPA routing, or 404 for unknown API routes
app.get('*', (req: express.Request, res: express.Response) => {
  if (req.path.startsWith('/api')) {
    res.status(404).json({ error: 'Not found', path: req.path });
    return;
  }
  res.sendFile(join(publicPath, 'index.html'));
});

// Start server
app.listen(config.port, () => {
  console.log(`Web Dot server running on http://localhost:${config.port}`);
  console.log(`Kernel at ${config.kernelUrl}`);
  console.log(`Namespace: ${config.namespace}`);
});
