# Web Dot - Product Ledger Demo UI

A tabbed web UI that demonstrates a single Product Ledger namespace through three primary views: Products, Intent, and Ledger.

## Architecture

```
Browser → Web Dot (Node/TS :3000) → Kernel (Go :8080)
```

- Browser only calls `http://localhost:3000/api/...` (same origin, no CORS)
- Web Dot serves UI from `/` and proxies `/api/*` → kernel `/v1/*`
- Namespace `ProductLedger:/Kesteron` is automatically injected by server

## Setup

### Prerequisites

- Node.js 18+ and npm
- Kernel server running on `http://localhost:8080`

### Installation

```bash
cd web-dot
npm install
```

### Development

```bash
# Start Web Dot server (serves UI + proxies API)
npm run dev

# Or start server only
npm run server
```

The server will start on `http://localhost:3000`.

### Production Build

```bash
npm run build
npm start
```

## Configuration

Edit `server/config.ts` to change:
- Kernel URL (default: `http://localhost:8080`)
- Namespace (default: `ProductLedger:/Kesteron`)
- Server port (default: `3000`)

Or set environment variables:
- `KERNEL_URL` - Kernel server URL
- `PORT` - Web Dot server port

## Project Structure

```
web-dot/
├── server/           # Node.js/TypeScript server
│   ├── index.ts     # Express server entry point
│   ├── config.ts    # Server configuration
│   └── routes/
│       └── api.ts   # API proxy routes
├── public/          # Static files (served from /)
│   └── index.html
├── src/             # Frontend source
│   ├── adapters/   # API client, GraphStore, Adapter
│   ├── components/ # React components (to be built)
│   ├── config/     # Demo configuration
│   └── types/      # TypeScript types
└── package.json
```

## API Endpoints

All endpoints are proxied to kernel with namespace automatically injected:

- `GET /api/healthz` → `GET /v1/healthz`
- `GET /api/expand?ids=...&depth=...` → `GET /v1/expand?...&namespace_id=...`
- `GET /api/history?target=...&limit=...` → `GET /v1/history?...&namespace_id=...`
- `GET /api/diff?a_seq=...&b_seq=...&target=...` → `GET /v1/diff?...&namespace_id=...`

## Status

- ✅ Phase 1: Project setup and server infrastructure
- ✅ Phase 2: Adapter layer (API client, GraphStore, Adapter)
- 🚧 Phase 3: UI foundation (in progress)
- ⏳ Phase 4-7: UI components (Products, Intent, Ledger tabs)
- ⏳ Phase 8-10: Styling, testing, documentation
