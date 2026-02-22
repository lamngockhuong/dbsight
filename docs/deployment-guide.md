# DBSight Deployment Guide

This guide covers deploying DBSight in development, staging, and production environments.

## Prerequisites

- Docker 20.10+ & Docker Compose v2+ (for containerized deployments)
- Go 1.26+ (for source builds)
- Node.js 20+ (for frontend build; not needed at runtime)
- PostgreSQL 14+ (app database for metadata/metrics storage)
- Target PostgreSQL instances with `pg_stat_statements` extension enabled

## Local Development

### Quick Start with Docker Compose

```bash
# Start PostgreSQL (app database)
docker-compose up -d postgres

# Generate encryption key
export ENCRYPTION_KEY=$(openssl rand -hex 32)
export DATABASE_URL="postgres://dbanalyzer:secret@localhost:5432/dbanalyzer?sslmode=disable"

# Run migrations
go run . migrate

# Start backend server (port 8080)
go run . serve
```

In another terminal:

```bash
cd web
npm install
npm run dev    # Frontend dev server on http://localhost:5173
# Vite proxy routes /api/* to http://localhost:8080
```

### Development without Docker

If PostgreSQL is already running locally:

```bash
# Create app database
createdb -U postgres dbanalyzer

# Set environment
export DATABASE_URL="postgres://postgres:password@localhost:5432/dbanalyzer"
export ENCRYPTION_KEY=$(openssl rand -hex 32)
export PORT=8080

# Run migrations and start server
go run . migrate && go run . serve
```

## Docker Deployments

### Build Image

```bash
# Multi-stage build (Node + Go + Alpine runtime)
docker build -t dbsight:latest .
```

Image size: ~60-80MB (includes Go binary ~15MB, React dist ~500KB, Alpine base ~5MB)

### Local Docker Compose

```bash
docker-compose up -d
# App available at http://localhost:8080
# PostgreSQL at localhost:5432
```

Edit `docker-compose.yml` to change ports, encryption key, or mount volumes for development.

### Run as Docker Container

```bash
docker run -d \
  --name dbsight \
  -e DATABASE_URL="postgres://user:pass@postgres-host:5432/dbanalyzer" \
  -e ENCRYPTION_KEY="$(openssl rand -hex 32)" \
  -e PORT=8080 \
  -p 8080:8080 \
  dbsight:latest serve
```

### Environment Variables

| Variable               | Required | Default | Example                                                          |
| ---------------------- | -------- | ------- | ---------------------------------------------------------------- |
| `DATABASE_URL`         | Yes      | —       | `postgres://user:pass@localhost:5432/dbanalyzer?sslmode=disable` |
| `ENCRYPTION_KEY`       | Yes      | —       | `abc123...` (64 hex chars = 32 bytes)                            |
| `PORT`                 | No       | `8080`  | `8080`                                                           |
| `WORKER_INTERVAL_SECS` | No       | `30`    | `30`                                                             |

Generate encryption key:

```bash
# openssl
openssl rand -hex 32

# or Go
go run -c 'package main; import ("crypto/rand"; "fmt"; "encoding/hex") func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b)) }'
```

## Production Deployment

### Pre-Deployment Checklist

- [ ] `pg_stat_statements` extension enabled on target databases
- [ ] PostgreSQL 14+ for app database (separate from target DBs)
- [ ] Encryption key generated and stored securely
- [ ] Database backups configured
- [ ] Firewall rules: allow 8080 (or reverse proxy port) inbound
- [ ] SSL/TLS certificate provisioned (via reverse proxy)

### Database Preparation

Create app database on PostgreSQL 14+:

```bash
createdb dbsight
# Or if using managed PostgreSQL (AWS RDS, etc.):
# Use console to create database, set admin user credentials
```

Enable `pg_stat_statements` on **each target database** you'll monitor:

```sql
-- On each target database:
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Verify
SELECT * FROM pg_stat_statements LIMIT 1;
```

### Reverse Proxy Setup (Nginx)

```nginx
# /etc/nginx/sites-available/dbsight

upstream dbsight {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name dbsight.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name dbsight.example.com;

    ssl_certificate /etc/letsencrypt/live/dbsight.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dbsight.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://dbsight;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # For SSE
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

Enable:

```bash
sudo ln -s /etc/nginx/sites-available/dbsight /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Systemd Service (Linux)

```ini
# /etc/systemd/system/dbsight.service

[Unit]
Description=DBSight Database Performance Analyzer
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=dbsight
WorkingDirectory=/opt/dbsight
ExecStart=/opt/dbsight/bin/dbsight serve
Restart=on-failure
RestartSec=5s

Environment="DATABASE_URL=postgres://user:pass@postgres.internal:5432/dbsight?sslmode=require"
Environment="ENCRYPTION_KEY=abc123..."
Environment="PORT=8080"
Environment="WORKER_INTERVAL_SECS=30"

StandardOutput=journal
StandardError=journal
SyslogIdentifier=dbsight

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable dbsight
sudo systemctl start dbsight
sudo journalctl -u dbsight -f
```

### Docker Compose (Production-Like)

For staging or small-scale production, use a production-ready compose setup:

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: dbsight
      POSTGRES_USER: dbsight
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    restart: always
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dbsight"]
      interval: 10s
      timeout: 5s
      retries: 5

  dbsight:
    image: dbsight:latest
    environment:
      DATABASE_URL: postgres://dbsight:${DB_PASSWORD}@postgres:5432/dbsight
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      PORT: 8080
      WORKER_INTERVAL_SECS: 30
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    restart: always

volumes:
  pgdata:
```

Deploy:

```bash
export DB_PASSWORD=$(openssl rand -base64 32)
export ENCRYPTION_KEY=$(openssl rand -hex 32)

docker-compose up -d
docker-compose ps
docker-compose logs -f dbsight
```

### Database Backups

For PostgreSQL databases, set up automated backups:

```bash
# Daily pg_dump
0 2 * * * pg_dump -U dbsight dbsight | gzip > /backups/dbsight-$(date +\%Y\%m\%d).sql.gz

# WAL archiving (for continuous backups)
# Update postgresql.conf:
# archive_mode = on
# archive_command = 'cp %p /backups/wal_archive/%f'
```

Retention policy:

- Daily backups: keep 7 days
- Weekly snapshots: keep 4 weeks
- Monthly snapshots: keep 1 year

Restore:

```bash
psql -U dbsight dbsight < /backups/dbsight-20260221.sql
```

## Monitoring & Observability

### Health Check Endpoint

Currently, DBSight does not expose a `/health` endpoint. For production, add readiness/liveness checks via dependency verification:

```bash
# Simple check: API responds
curl -f http://localhost:8080/api/connections || exit 1
```

Post-MVP will add dedicated `/health` endpoint with:

- Database connectivity status
- Worker heartbeat (last collection timestamp)
- Migration status

### Structured Logging

Logs are output to stderr in JSON format (slog):

```text
{"time":"2026-02-21T12:00:00Z","level":"ERROR","msg":"worker stopped","err":"connection failed"}
```

Aggregate with:

```bash
# Docker: capture stderr
docker logs dbsight | grep -i error

# Systemd: journal
journalctl -u dbsight | grep ERROR

# Syslog integration: redirect to /dev/log
dbsight serve 2>&1 | logger -t dbsight
```

### Performance Tuning

**PostgreSQL Connection Pool**

Default pgxpool settings (tunable in future):

- Max conns: 10
- Min conns: 2

For high-traffic deployments, consider:

```go
// Future: add pgx pool config via env vars
config.MaxConns = 25    // Increase for many target DBs
config.MinConns = 5
```

**Worker Interval**

- `WORKER_INTERVAL_SECS=30` (default): balanced polling
- `WORKER_INTERVAL_SECS=10`: more frequent updates, higher load
- `WORKER_INTERVAL_SECS=60`: less frequent, lower load

**Frontend Caching**

Static assets (JS, CSS) are served from Go binary with far-future cache headers. No CDN needed for MVP.

## Kubernetes Deployment (Optional)

For container orchestration, use these manifests as templates:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dbsight-config
data:
  PORT: "8080"
  WORKER_INTERVAL_SECS: "30"
---
apiVersion: v1
kind: Secret
metadata:
  name: dbsight-secrets
type: Opaque
stringData:
  DATABASE_URL: postgres://user:pass@postgres-svc:5432/dbsight
  ENCRYPTION_KEY: abc123...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dbsight
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dbsight
  template:
    metadata:
      labels:
        app: dbsight
    spec:
      containers:
        - name: dbsight
          image: dbsight:latest
          ports:
            - containerPort: 8080
          envFrom:
            - configMapRef:
                name: dbsight-config
            - secretRef:
                name: dbsight-secrets
          resources:
            requests:
              cpu: 250m
              memory: 128Mi
            limits:
              cpu: 1000m
              memory: 512Mi
          livenessProbe:
            httpGet:
              path: /api/connections
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 30
---
apiVersion: v1
kind: Service
metadata:
  name: dbsight-svc
spec:
  type: ClusterIP
  selector:
    app: dbsight
  ports:
    - port: 80
      targetPort: 8080
```

Deploy:

```bash
kubectl apply -f k8s-manifest.yaml
kubectl port-forward svc/dbsight-svc 8080:80
```

## Troubleshooting

### Connection Refused

```text
error: connection failed: invalid DSN
```

- Verify target database credentials and network connectivity
- Check `pg_stat_statements` is installed: `psql -c "SELECT count(*) FROM pg_stat_statements;"`
- Ensure firewall allows connection from DBSight host

### Worker Not Collecting Metrics

```bash
journalctl -u dbsight | grep worker
```

Check:

- Database reachability (run test endpoint manually)
- `pg_stat_statements` extension enabled
- Worker logs for panics or timeouts

### SSL/TLS Certificate Issues

With self-signed certificates on target databases:

```bash
# Disable SSL verification (dev only)
DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=disable"

# Or use root CA (production)
DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=require&sslrootcert=/path/to/ca.crt"
```

### Encryption Key Mismatch

Old encrypted DSNs fail after key rotation. Current workaround:

- Regenerate connections in UI (re-enter credentials)
- Re-encrypt stored data (not yet automated)

Future: implement key rotation migration.

### High Memory Usage

Check worker concurrency and connection pooling:

```bash
# Reduce concurrent collectors
# Future: add WORKER_MAX_CONCURRENCY env var

# Or increase interval
WORKER_INTERVAL_SECS=60
```

## Upgrading DBSight

1. Pull latest code / update Docker image
2. Run migrations: `go run . migrate` (idempotent)
3. Restart service
4. No data loss (schema backward compatible)

Breaking changes will be documented in releases.

## Disaster Recovery

### Database Loss

If app database is lost but target DBs intact:

1. Restore from backup: `psql -U dbsight dbsight < backup.sql`
2. Restart DBSight: `systemctl restart dbsight`
3. Re-test connections in UI

If no backup available:

1. Recreate database schema: `go run . migrate`
2. Re-add connections in UI (credentials will need re-entry)
3. Query history is lost, but collection resumes

### Encryption Key Loss

If `ENCRYPTION_KEY` is lost, stored encrypted DSNs cannot be decrypted. Mitigation:

- Store encryption key in secure vault (1Password, HashiCorp Vault, AWS Secrets Manager)
- Rotate key periodically (requires data re-encryption; not automated yet)
- Never commit `.env` or secrets to version control

---

**Document Version:** 1.0
**Last Updated:** 2026-02-21
