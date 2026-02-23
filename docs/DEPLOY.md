# Deploying ctrldot.dev

This doc describes how to deploy the minimal static site to **ctrldot.dev** via GitHub Pages and how to configure DNS.

## How to deploy (GitHub Pages)

1. **Repo:** Ensure the repo is at `https://github.com/ctrldot-dev/ctrldot` (or update the workflow and links in `site/` for your org/repo).

2. **Pages source:** In the repo, go to **Settings → Pages**. Set:
   - **Source:** GitHub Actions (not “Deploy from a branch”).

3. **Workflow:** The workflow `.github/workflows/pages.yml` runs on push to `main`. It uploads the `site/` folder and deploys it to GitHub Pages. No build step — static files only.

4. **Push to main:** After merging to `main`, the workflow runs. When it finishes, the site is available at `https://<owner>.github.io/ctrldot` (or your Pages URL).

5. **Custom domain:** In **Settings → Pages**, under “Custom domain”, enter `ctrldot.dev` and save. Enable **Enforce HTTPS** when offered.

## DNS configuration for ctrldot.dev

After you add the custom domain in GitHub, GitHub will show the target host (e.g. `ctrldot-dev.github.io`). Use the records below at your DNS registrar. Replace `ctrldot-dev.github.io` with the value shown in GitHub Pages settings if different.

### Root domain (ctrldot.dev)

GitHub Pages supports custom root domains via A records. Use GitHub’s documented IPs (they can change; check [GitHub Pages docs](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site)):

| Type | Name/Host | Value | TTL (optional) |
|------|-----------|--------|----------------|
| A    | `@`       | `185.199.108.153` | 3600 |
| A    | `@`       | `185.199.109.153` | 3600 |
| A    | `@`       | `185.199.110.153` | 3600 |
| A    | `@`       | `185.199.111.153` | 3600 |

*(Confirm current IPs in [GitHub’s documentation](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site) — they may update.)*

### www subdomain (www.ctrldot.dev)

| Type  | Name/Host | Value                 | TTL (optional) |
|-------|------------|------------------------|----------------|
| CNAME | `www`      | `ctrldot-dev.github.io` | 3600         |

Use the exact CNAME target shown in your repo’s Pages settings (e.g. `ctrldot-dev.github.io`).

### Copy-paste checklist

At your registrar’s DNS panel:

1. Add the four **A** records for `@` (root) with the IPs above (or the latest from GitHub).
2. Add the **CNAME** record for `www` pointing to `ctrldot-dev.github.io` (or your Pages host).
3. Wait for propagation (up to 48 hours, often minutes).
4. In GitHub Pages settings, set custom domain to `ctrldot.dev` and enable **Enforce HTTPS**.

## Post-deploy checks

- Open `https://ctrldot.dev` — loads without certificate warnings.
- Test `https://www.ctrldot.dev` if you use www.
- Verify links: GitHub repo link, mailto `contact@ctrldot.dev`.
- Optional: add `site/.well-known/security.txt` (see SETUP_GUIDE.md) and ensure it’s served at `https://ctrldot.dev/.well-known/security.txt`.

## Email (contact@ctrldot.dev)

For deliverability, add SPF, DKIM, and DMARC. The exact records depend on your email provider (e.g. Google Workspace, Proton, Fastmail). Ask your provider for:

- **SPF:** e.g. `v=spf1 include:_spf.example.com ~all` (replace with your provider’s include).
- **DKIM:** provider gives a host (e.g. `selector._domainkey`) and a TXT value.
- **DMARC:** e.g. `v=DMARC1; p=none; rua=mailto:dmarc@ctrldot.dev` to start (no policy enforcement). Tighten to `p=quarantine` or `p=reject` later.

If you share which provider you use, these can be filled in with exact values.
