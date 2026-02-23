# Deploy Decision Ledger Demo to Hetzner

Deploy the Decision Ledger demo so it is available at **https://www.garethapi.com/decision-ledger-demo/**.

**Two servers:**

| Role | Server | Notes |
|------|--------|--------|
| **garethapi.com** (nginx, main site) | **91.99.203.148** | Add nginx `location` to proxy `/decision-ledger-demo/` to the app server |
| **Decision Ledger app** (Postgres, kernel, web-dot) | **89.167.16.171** | Deploy target; install path `/opt/decision-ledger-demo` |

**Security:** Do not commit server passwords. Use SSH keys (e.g. `ssh-copy-id root@89.167.16.171`).

---

## 1. Prerequisites on your machine

- SSH access to the Decision Ledger app server (89.167.16.171).
- Optional: set `DLD_SERVER=root@89.167.16.171` so you can copy-paste commands below.

---

## 2. One-time setup on the Decision Ledger app server (89.167.16.171)

SSH into the **app server** and install dependencies, then create services.

```bash
ssh root@89.167.16.171
```

### 2.1 Install Node.js (LTS), Go, Docker

```bash
# Node 20 (adjust for your distro)
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y nodejs

# Go (e.g. 1.21)
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

# Docker (for Postgres)
apt-get update && apt-get install -y ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a644 /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update && apt-get install -y docker-ce docker-ce-cli containerd.io
```

### 2.2 Directory and app config

The app directory should already exist: `/opt/decision-ledger-demo`. Create env files there **on the server** (do not commit secrets). Use the same DB password as in `docker-compose.yml` (e.g. `kernel`) or set a stronger one and pass it to Postgres.

- **`/opt/decision-ledger-demo/.env.kernel`** — e.g.  
  `DB_URL=postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable`  
  `PORT=8080`
- **`/opt/decision-ledger-demo/.env.webdot`** — e.g.  
  `KERNEL_URL=http://127.0.0.1:8080`  
  `PORT=3000`

Create them before the first start of the systemd units (or restart after creating).

### 2.3 Reverse proxy (nginx) on the garethapi.com server (91.99.203.148)

Nginx that serves **www.garethapi.com** runs on **91.99.203.148**. The Decision Ledger app runs on **89.167.16.171**. On the **garethapi.com server** (91.99.203.148), add a location that proxies `/decision-ledger-demo/` to the app server.

**On 91.99.203.148** — inside the `server { ... }` block for `www.garethapi.com` (and for `garethapi.com` if you have a separate block):

```nginx
location /decision-ledger-demo/ {
    auth_basic "Decision Ledger Demo Login";
    auth_basic_user_file /etc/nginx/.htpasswd-dld;
    proxy_pass http://89.167.16.171:3000/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Create the password file **once** on 91.99.203.148 (see [§ 2.3.1 Securing the demo with login](#231-securing-the-demo-with-login)).

So:

- Request: `https://www.garethapi.com/decision-ledger-demo/` → browser is prompted for username/password, then proxy to `http://89.167.16.171:3000/`
- Request: `https://www.garethapi.com/decision-ledger-demo/api/config` → same auth, then proxy to `http://89.167.16.171:3000/api/config`

The frontend uses the browser path (`/decision-ledger-demo`) for API and assets; no extra base-path config is needed on the app. On **89.167.16.171**, web-dot must listen on port 3000 (all interfaces by default). For security, consider allowing inbound port 3000 on 89.167.16.171 only from 91.99.203.148.

Reload nginx **on 91.99.203.148** after editing:

```bash
nginx -t && systemctl reload nginx
```

#### 2.3.1 Securing the demo with login

Use **HTTP Basic Authentication** on the nginx server (91.99.203.148) so visitors must enter a username and password before accessing the demo.

1. **SSH to the garethapi.com server:**
   ```bash
   ssh root@91.99.203.148
   ```

2. **Install `apache2-utils`** (for `htpasswd`):
   ```bash
   apt-get update && apt-get install -y apache2-utils
   ```

3. **Create the password file** (choose a strong password when prompted):
   ```bash
   htpasswd -c /etc/nginx/.htpasswd-dld demo
   ```
   To add another user later (without overwriting the file), omit `-c`:
   ```bash
   htpasswd /etc/nginx/.htpasswd-dld anotheruser
   ```

4. **Use the nginx location block** from § 2.3 above (it includes `auth_basic` and `auth_basic_user_file`). If you already had a location without auth, add these two lines inside the same `location /decision-ledger-demo/ { ... }` block:
   ```nginx
   auth_basic "Decision Ledger Demo Login";
   auth_basic_user_file /etc/nginx/.htpasswd-dld;
   ```

5. **Test and reload nginx:**
   ```bash
   nginx -t && systemctl reload nginx
   ```

Visiting **https://www.garethapi.com/decision-ledger-demo/** (or **https://garethapi.com/decision-ledger-demo/**) will now prompt for the username and password you set.

#### 2.3.2 Alternative: Microsoft Entra ID (oauth2-proxy)

For SSO with Microsoft Entra ID (Azure AD) instead of Basic Auth, use **oauth2-proxy** behind nginx and protect `/decision-ledger-demo/` with `auth_request`. This gives a redirect to Microsoft sign-in and optional email allowlisting.

- **Guide:** [DEPLOY-HETZNER-ENTRA-OAUTH2PROXY.md](./DEPLOY-HETZNER-ENTRA-OAUTH2PROXY.md)
- **Templates:** `scripts/decision-ledger-auth/` (Docker Compose, `.env.example`, `allowed_emails.txt.example`)

Setup is on the **garethapi.com server** (91.99.203.148): run oauth2-proxy in Docker, add nginx locations for `/oauth2/`, and switch the `/decision-ledger-demo/` block to use `auth_request` instead of `auth_basic`. If you switch to Entra, remove the Basic Auth directives from that location.

### 2.4 Systemd units (on 89.167.16.171)

Create two services so the kernel and web-dot run on boot and after deploy.

**Kernel** — `/etc/systemd/system/decision-ledger-kernel.service`:

```ini
[Unit]
Description=Decision Ledger Kernel API
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
WorkingDirectory=/opt/decision-ledger-demo
EnvironmentFile=/opt/decision-ledger-demo/.env.kernel
ExecStart=/opt/decision-ledger-demo/kernel
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

**Web Dot** — `/etc/systemd/system/decision-ledger-webdot.service`:

```ini
[Unit]
Description=Decision Ledger Web UI
After=network.target decision-ledger-kernel.service
Requires=decision-ledger-kernel.service

[Service]
Type=simple
WorkingDirectory=/opt/decision-ledger-demo/web-dot
EnvironmentFile=/opt/decision-ledger-demo/.env.webdot
ExecStart=/usr/bin/node dist/server/index.js
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Then:

```bash
systemctl daemon-reload
systemctl enable decision-ledger-kernel decision-ledger-webdot
```

---

## 3. Deploy from your machine

From the repo root (futurematic-kernal):

1. **Sync files** to the **Decision Ledger app server** (excludes .git, node_modules, etc.):

```bash
export DLD_SERVER="${DLD_SERVER:-root@89.167.16.171}"
rsync -avz --delete \
  --exclude=.git \
  --exclude=node_modules \
  --exclude=web-dot/node_modules \
  --exclude=web-dot/dist \
  --exclude=.env.kernel \
  --exclude=.env.webdot \
  --exclude=*.db \
  ./ "$DLD_SERVER:/opt/decision-ledger-demo/"
```

You will be prompted for the root password unless you use SSH keys.

2. **On 89.167.16.171**: build, migrate, and start:

```bash
ssh "$DLD_SERVER" "cd /opt/decision-ledger-demo && \
  docker compose up -d && \
  sleep 3 && \
  for f in migrations/0001_init.sql migrations/0002_finledger_namespaces.sql migrations/0003_productledger_namespaces.sql; do \
    docker compose exec -T postgres psql -U kernel -d kernel -f /docker-entrypoint-initdb.d/\$(basename \$f) 2>/dev/null || true; \
  done && \
  (test -f cmd/kernel/main.go && /usr/local/go/bin/go build -o kernel cmd/kernel/main.go || true) && \
  cd web-dot && npm ci && npm run build && cd .. && \
  systemctl restart decision-ledger-kernel decision-ledger-webdot"
```

If your kernel entry point is not `cmd/kernel/main.go`, build the kernel binary your usual way and place it at `/opt/decision-ledger-demo/kernel` on **89.167.16.171**.

---

## 4. First-time data setup (on 89.167.16.171)

After Postgres and kernel are running:

- Run bootstrap for Fin and Product ledger policy sets (see repo `Makefile` / `SEED_PRODUCT_LEDGER.md`).
- Run migrations 0004/0005 only if you need a clean or Fin-only wipe.
- Seed Product Ledger and Fin Ledger (e.g. `go run ./cmd/seed_finledger ...`, product seed scripts).
- Optionally run the materials seed script for the Product namespaces.

---

## 5. Optional: deploy script

Use the script from the repo (reads `DLD_SERVER`, does not use passwords from disk):

```bash
./scripts/deploy-to-hetzner.sh
```

You will be prompted for SSH password unless you use key-based auth.

---

## 6. Verify

- Open **https://www.garethapi.com/decision-ledger-demo/** (or **https://garethapi.com/decision-ledger-demo/**) — the Decision Ledger demo UI should load (full app with left nav, tabs, etc.).
- Check services: `ssh root@89.167.16.171 'systemctl status decision-ledger-kernel decision-ledger-webdot'`.

---

## 7. Troubleshooting

### You see "Decision Ledger Demo OK" instead of the full UI

That text is **not** from the Decision Ledger app. It comes from a **test/placeholder response** on the **garethapi.com server (91.99.203.148)** — nginx is serving a simple "OK" page instead of proxying to the app.

**Fix on 91.99.203.148:**

1. SSH in: `ssh root@91.99.203.148`
2. Find the nginx config for garethapi.com (and www if you use it), e.g.:
   ```bash
   grep -r "decision-ledger-demo" /etc/nginx/
   ```
3. For `/decision-ledger-demo/` you must **proxy** to the app server. Remove any static response for that path, e.g.:
   - `return 200 'Decision Ledger Demo OK';` or similar
   - `root /some/path;` or `alias /some/file;` that serves a test page
4. Replace (or add) the location with the proxy block (include auth if you use login; create `/etc/nginx/.htpasswd-dld` first — see § 2.3.1):
   ```nginx
   location /decision-ledger-demo/ {
       auth_basic "Decision Ledger Demo Login";
       auth_basic_user_file /etc/nginx/.htpasswd-dld;
       proxy_pass http://89.167.16.171:3000/;
       proxy_http_version 1.1;
       proxy_set_header Host $host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_set_header X-Forwarded-Proto $scheme;
   }
   ```
5. Apply the same in **both** server blocks if you have separate configs for `garethapi.com` and `www.garethapi.com` (you’re visiting https://garethapi.com/… so that server block must proxy too).
6. Reload nginx: `nginx -t && systemctl reload nginx`
