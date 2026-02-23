# Decision Ledger Demo — Microsoft Entra ID login (nginx + oauth2-proxy)

This adds **Microsoft Entra ID (Azure AD)** login in front of the demo using [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/) behind nginx. Use this instead of Basic Auth if you want SSO with your organisation.

**Assumptions:**

- Demo is already reachable at `https://garethapi.com/decision-ledger-demo/` (see [DEPLOY-HETZNER.md](./DEPLOY-HETZNER.md)).
- nginx runs on **91.99.203.148** (garethapi.com) and proxies `/decision-ledger-demo/` to the app on **89.167.16.171**.
- You have an Entra ID **App registration** with:
  - **Tenant ID**, **Client ID**, **Client secret**
  - Redirect URI: `https://<your-domain>/oauth2/callback` (e.g. `https://garethapi.com/oauth2/callback`).

**Do not commit** tenant ID, client ID, client secret, or cookie secret. Keep them only on the server or in a local `.env` used on the server.

---

## Overview

1. Run **oauth2-proxy** in Docker on the **garethapi.com server** (91.99.203.148).
2. Add nginx locations for `/oauth2/` and use `auth_request` to protect `/decision-ledger-demo/`.
3. Optionally restrict access with an **email allowlist**.

---

## 1. Prerequisites on the garethapi.com server (91.99.203.148)

- Docker (and Docker Compose) installed.
- Nginx config for garethapi.com (and www if used) that you can edit.

**Entra App registration:**

- In Azure Portal → Microsoft Entra ID → App registrations → your app (or create one):
  - **Authentication** → Add a redirect URI: `https://<your-domain>/oauth2/callback` (e.g. `https://garethapi.com/oauth2/callback`).
  - **Certificates & secrets** → New client secret; copy the **Value** (this is `CLIENT_SECRET`).
  - Note **Application (client) ID** and **Directory (tenant) ID**.

---

## 2. One-time setup on 91.99.203.148

SSH in:

```bash
ssh root@91.99.203.148
```

Set variables (use your real values; do not commit these):

```bash
export DOMAIN="garethapi.com"
export TENANT_ID="YOUR_TENANT_ID"
export CLIENT_ID="YOUR_CLIENT_ID"
export CLIENT_SECRET="YOUR_CLIENT_SECRET"
```

---

## 3. Create auth directory

```bash
sudo mkdir -p /opt/decision-ledger-auth
sudo chown -R "$USER":"$USER" /opt/decision-ledger-auth
cd /opt/decision-ledger-auth
```

---

## 4. Cookie secret for oauth2-proxy

Generate a strong cookie secret (oauth2-proxy needs 16, 24, or 32 decoded bytes; 16 bytes → 24-char base64):

```bash
python3 - << 'PY'
import os, base64
print(base64.b64encode(os.urandom(16)).decode())
PY
```

Set it:

```bash
export COOKIE_SECRET="PASTE_THE_OUTPUT_HERE"
```

---

## 5. Create `.env` for Docker Compose

Create `/opt/decision-ledger-auth/.env` (only on the server, never commit):

```bash
cat > /opt/decision-ledger-auth/.env <<EOF
DOMAIN=${DOMAIN}
TENANT_ID=${TENANT_ID}
CLIENT_ID=${CLIENT_ID}
CLIENT_SECRET=${CLIENT_SECRET}
COOKIE_SECRET=${COOKIE_SECRET}
EOF
```

Verify:

```bash
ls -la /opt/decision-ledger-auth/.env
```

---

## 6. Email allowlist (recommended)

Restrict who can sign in. Create `/opt/decision-ledger-auth/allowed_emails.txt`:

```bash
cat > /opt/decision-ledger-auth/allowed_emails.txt <<'EOF'
# One email per line. Lines starting with # are comments.
your.email@example.com
EOF
```

Add more lines for other allowed users. This file is mounted into the oauth2-proxy container.

---

## 7. Docker Compose for oauth2-proxy

Copy the template from the repo (or create the file manually). From your **local** machine, sync the auth template to the server:

```bash
rsync -avz scripts/decision-ledger-auth/ root@91.99.203.148:/opt/decision-ledger-auth/
```

Then on **91.99.203.148** ensure `.env` and `allowed_emails.txt` exist (they are not in the repo). Start oauth2-proxy:

```bash
cd /opt/decision-ledger-auth
docker compose -f docker-compose.auth.yml up -d
docker compose -f docker-compose.auth.yml ps
```

Check logs:

```bash
docker compose -f docker-compose.auth.yml logs -f --tail=200
```

There should be no fatal errors. The container listens on port 4180.

---

## 8. Nginx: oauth2-proxy endpoints and protect `/decision-ledger-demo/`

All of this is on **91.99.203.148**.

### 8.1 Find nginx site config

```bash
sudo grep -R "server_name garethapi.com" -n /etc/nginx/sites-enabled /etc/nginx/conf.d 2>/dev/null | head -20
```

Set the path (example):

```bash
export NGINX_SITE="/etc/nginx/sites-enabled/garethapi.com"
```

### 8.2 Back up

```bash
sudo cp "$NGINX_SITE" "${NGINX_SITE}.bak.$(date +%Y%m%d%H%M%S)"
```

### 8.3 Add oauth2-proxy locations

Inside the same `server { ... }` block that serves garethapi.com (and www if separate), add:

```nginx
# --- oauth2-proxy ---
# Use ^~ so this wins over regex locations (e.g. static .css) that would otherwise match /oauth2/static/*
location ^~ /oauth2/ {
  proxy_pass http://127.0.0.1:4180;
  proxy_set_header Host $host;
  proxy_set_header X-Real-IP $remote_addr;
  proxy_set_header X-Scheme $scheme;
  proxy_set_header X-Forwarded-Proto $scheme;
  proxy_set_header X-Forwarded-Host $host;
  proxy_set_header X-Forwarded-Uri $request_uri;
  proxy_buffer_size 128k;
  proxy_buffers 4 256k;
  proxy_busy_buffers_size 256k;
  proxy_temp_file_write_size 256k;
}

location = /oauth2/auth {
  internal;
  proxy_pass http://127.0.0.1:4180;
  proxy_set_header Host $host;
  proxy_set_header X-Real-IP $remote_addr;
  proxy_set_header X-Scheme $scheme;
  proxy_set_header X-Forwarded-Proto $scheme;
  proxy_set_header X-Forwarded-Host $host;
  proxy_set_header X-Forwarded-Uri $request_uri;
  proxy_set_header Content-Length "";
  proxy_pass_request_body off;
}
# --- end oauth2-proxy ---
```

### 8.4 Protect `/decision-ledger-demo/`

**If you currently use Basic Auth** for `/decision-ledger-demo/`, remove the `auth_basic` and `auth_basic_user_file` lines and use the block below instead.

Replace (or merge into) your existing `location /decision-ledger-demo/` block with:

```nginx
location /decision-ledger-demo/ {
  auth_request /oauth2/auth;
  error_page 401 = /oauth2/sign_in;

  auth_request_set $user  $upstream_http_x_auth_request_user;
  auth_request_set $email $upstream_http_x_auth_request_email;
  proxy_set_header X-User  $user;
  proxy_set_header X-Email $email;

  proxy_pass http://89.167.16.171:3000/;
  proxy_http_version 1.1;
  proxy_set_header Host $host;
  proxy_set_header X-Real-IP $remote_addr;
  proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  proxy_set_header X-Forwarded-Proto $scheme;
}
```

### 8.5 Test and reload

```bash
sudo nginx -t
sudo systemctl reload nginx
```

If `nginx -t` fails, restore the backup and fix:

```bash
sudo cp "${NGINX_SITE}.bak.TIMESTAMP" "$NGINX_SITE"
sudo nginx -t && sudo systemctl reload nginx
```

---

## 9. Test end-to-end

1. In a private/incognito window, open `https://garethapi.com/decision-ledger-demo/`.
2. You should be redirected to Microsoft sign-in.
3. After sign-in, you should land on the demo.

If you see a permission/forbidden error:

- Ensure your email is in `/opt/decision-ledger-auth/allowed_emails.txt`.
- Restart oauth2-proxy:  
  `docker compose -f /opt/decision-ledger-auth/docker-compose.auth.yml restart`  
  and check logs.

---

## 10. Optional: show “Signed in as” in the UI

nginx forwards `X-User` and `X-Email` to the upstream app. The Web Dot app can read these and show “Signed in as …” without implementing auth itself. No change is required for the login gate to work.

---

## 11. Operational commands

| Task | Command |
|------|---------|
| Logs | `docker compose -f /opt/decision-ledger-auth/docker-compose.auth.yml logs -f --tail=200` |
| Stop | `docker compose -f /opt/decision-ledger-auth/docker-compose.auth.yml down` |
| Start | `docker compose -f /opt/decision-ledger-auth/docker-compose.auth.yml up -d` |
| Update allowlist | Edit `allowed_emails.txt`, then `docker compose -f /opt/decision-ledger-auth/docker-compose.auth.yml restart` |

---

## 12. Security notes

- **No Basic Auth** when using this flow; auth is OIDC via Entra.
- **Secure cookies**: oauth2-proxy is configured with `cookie-secure=true`, `cookie-samesite=lax`.
- **Allowlist** limits sign-in to listed emails.
- **Auth at the edge (nginx)** keeps the app simple; secrets stay on the nginx server.
