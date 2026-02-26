import fs from "node:fs/promises";
import path from "node:path";

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";

type Doc = {
  id: string;          // stable id (path-based)
  title: string;       // inferred from first H1, or filename
  filePath: string;    // absolute path
  relPath: string;     // path relative to docs root
  text: string;        // full markdown text
  headings: { level: number; text: string; line: number }[];
};

const DOCS_ROOT = process.env.CTRLDOT_DOCS_ROOT
  ? path.resolve(process.env.CTRLDOT_DOCS_ROOT)
  : path.resolve(process.cwd(), "..", "docs");

const MAX_DOC_BYTES = Number(process.env.CTRLDOT_MAX_DOC_BYTES ?? 600_000); // ~600KB
const MAX_SEARCH_RESULTS = Number(process.env.CTRLDOT_MAX_SEARCH_RESULTS ?? 10);

function normalise(s: string) {
  return s.toLowerCase();
}

function fileIdFromRelPath(relPath: string) {
  // URI-safe-ish stable id
  return relPath.replace(/\\/g, "/");
}

function inferTitle(relPath: string, text: string) {
  const lines = text.split(/\r?\n/);
  for (const line of lines) {
    const m = /^#\s+(.+?)\s*$/.exec(line);
    if (m) return m[1].trim();
  }
  // fallback: filename
  const base = path.basename(relPath).replace(/\.md$/i, "");
  return base.replace(/[-_]/g, " ");
}

function extractHeadings(text: string) {
  const headings: Doc["headings"] = [];
  const lines = text.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const m = /^(#{2,6})\s+(.+?)\s*$/.exec(line);
    if (!m) continue;
    headings.push({
      level: m[1].length,
      text: m[2].trim(),
      line: i + 1,
    });
  }
  return headings;
}

async function* walk(dir: string): AsyncGenerator<string> {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  for (const e of entries) {
    const full = path.join(dir, e.name);
    if (e.isDirectory()) yield* walk(full);
    else if (e.isFile() && e.name.toLowerCase().endsWith(".md")) yield full;
  }
}

async function loadDocs(): Promise<Doc[]> {
  const docs: Doc[] = [];
  for await (const absPath of walk(DOCS_ROOT)) {
    const relPath = path.relative(DOCS_ROOT, absPath);
    const stat = await fs.stat(absPath);
    if (stat.size > MAX_DOC_BYTES) continue; // skip huge files for safety
    const text = await fs.readFile(absPath, "utf8");
    docs.push({
      id: fileIdFromRelPath(relPath),
      title: inferTitle(relPath, text),
      filePath: absPath,
      relPath,
      text,
      headings: extractHeadings(text),
    });
  }
  // stable ordering
  docs.sort((a, b) => a.relPath.localeCompare(b.relPath));
  return docs;
}

function snippetAround(text: string, idx: number, radius = 120) {
  const start = Math.max(0, idx - radius);
  const end = Math.min(text.length, idx + radius);
  const s = text.slice(start, end);
  return (start > 0 ? "…" : "") + s.replace(/\s+/g, " ").trim() + (end < text.length ? "…" : "");
}

function toAnchor(heading: string) {
  // GitHub-ish anchor
  return heading
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, "")
    .replace(/\s+/g, "-");
}

async function main() {
  const server = new McpServer({
    name: "ctrldot-docs",
    version: "0.1.0",
  });

  let docs = await loadDocs();

  // Optional: reload docs on each call (simple + safe for early stage)
  const reloadIfRequested = async () => {
    if (process.env.CTRLDOT_DOCS_RELOAD === "1") {
      docs = await loadDocs();
    }
  };

  // TOOL: List Docs
  server.tool(
    "List Docs",
    "List all available documentation pages (read-only).",
    z.object({}),
    async () => {
      await reloadIfRequested();
      const items = docs.map((d) => ({
        id: d.id,
        title: d.title,
        path: d.relPath.replace(/\\/g, "/"),
      }));
      return {
        content: [
          {
            type: "text",
            text: JSON.stringify({ docsRoot: DOCS_ROOT, items }, null, 2),
          },
        ],
      };
    }
  );

  // TOOL: Fetch Doc
  server.tool(
    "Fetch Doc",
    "Fetch a single documentation page by id (read-only).",
    z.object({
      id: z.string().min(1),
      includeHeadings: z.boolean().optional().default(true),
    }),
    async ({ id, includeHeadings }) => {
      await reloadIfRequested();
      const doc = docs.find((d) => d.id === id);
      if (!doc) {
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify({ error: "DOC_NOT_FOUND", id }, null, 2),
            },
          ],
        };
      }
      const payload: any = {
        id: doc.id,
        title: doc.title,
        path: doc.relPath.replace(/\\/g, "/"),
        text: doc.text,
      };
      if (includeHeadings) {
        payload.headings = doc.headings.map((h) => ({
          level: h.level,
          text: h.text,
          line: h.line,
          anchor: toAnchor(h.text),
        }));
      }
      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(payload, null, 2),
          },
        ],
      };
    }
  );

  // TOOL: Search Docs
  server.tool(
    "Search Docs",
    "Search across documentation pages and return the best matches with snippets (read-only).",
    z.object({
      query: z.string().min(1),
      limit: z.number().int().min(1).max(50).optional().default(MAX_SEARCH_RESULTS),
    }),
    async ({ query, limit }) => {
      await reloadIfRequested();
      const q = normalise(query);
      const results: Array<{
        id: string;
        title: string;
        path: string;
        score: number;
        snippet: string;
        heading?: { text: string; anchor: string; line: number };
      }> = [];

      for (const d of docs) {
        const hay = normalise(d.text);
        const idx = hay.indexOf(q);
        if (idx === -1) continue;

        // Simple scoring: earlier hit + shorter doc wins, with title bonus
        const titleHit = normalise(d.title).includes(q) ? 50 : 0;
        const score =
          titleHit + Math.max(0, 10_000 - idx) + Math.max(0, 2_000 - d.text.length / 10);

        // Find nearest heading above the hit for better citations
        let nearest: Doc["headings"][number] | undefined;
        for (const h of d.headings) {
          const lineStartIdx = d.text.split(/\r?\n/).slice(0, h.line).join("\n").length;
          if (lineStartIdx <= idx) nearest = h;
          else break;
        }

        results.push({
          id: d.id,
          title: d.title,
          path: d.relPath.replace(/\\/g, "/"),
          score,
          snippet: snippetAround(d.text, idx),
          heading: nearest
            ? { text: nearest.text, anchor: toAnchor(nearest.text), line: nearest.line }
            : undefined,
        });
      }

      results.sort((a, b) => b.score - a.score);
      const top = results.slice(0, limit);

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify({ query, results: top }, null, 2),
          },
        ],
      };
    }
  );

  // Transport: stdio
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

