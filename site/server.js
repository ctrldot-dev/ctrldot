const path = require('path');
const fs = require('fs');
const express = require('express');
const { marked } = require('marked');

const app = express();
const siteDir = __dirname;
const docsDir = path.join(siteDir, '..', 'docs');
const PORT = process.env.PORT || 3001;

function slugFromFilename(name) {
  return name.replace(/\.md$/i, '').toLowerCase().replace(/_/g, '-');
}

function findMdBySlug(slug) {
  if (!fs.existsSync(docsDir)) return null;
  const files = fs.readdirSync(docsDir).filter((f) => f.endsWith('.md'));
  const name = files.find((f) => slugFromFilename(f) === slug);
  return name ? path.join(docsDir, name) : null;
}

function extractTitle(md) {
  const m = md.match(/^#\s+(.+)$/m);
  return m ? m[1].trim() : 'Documentation';
}

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function docLayout(title, bodyHtml) {
  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${escapeHtml(title)} — Ctrl Dot</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Azeret+Mono:wght@400;500&display=swap" rel="stylesheet">
  <link rel="stylesheet" href="/style.css">
</head>
<body>
  <header>
    <a href="/"><img src="/assets/logo.png" alt="Ctrl Dot" class="logo"></a>
    <hr>
  </header>

  <main class="doc-content">
    ${bodyHtml}
  </main>

  <footer>
    <nav class="bottom-nav">
      <a href="https://github.com/ctrldot-dev/ctrldot">[GitHub]</a>
      <a href="/docs.html">[Docs]</a>
      <a href="/contact.html">[Contact]</a>
    </nav>
  </footer>
</body>
</html>
`;
}

marked.setOptions({ gfm: true });

app.get('/docs/:slug', (req, res) => {
  const slug = req.params.slug.toLowerCase();
  const mdPath = findMdBySlug(slug);
  if (!mdPath) {
    return res.status(404).send('Doc not found');
  }
  const md = fs.readFileSync(mdPath, 'utf8');
  const title = extractTitle(md);
  const bodyHtml = marked.parse(md);
  res.send(docLayout(title, bodyHtml));
});

app.use(express.static(siteDir));

app.listen(PORT, '0.0.0.0', () => {
  console.log(`Site with dynamic docs listening on port ${PORT}`);
});
