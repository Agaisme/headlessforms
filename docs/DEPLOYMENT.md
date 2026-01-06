# HeadlessForms Deployment Guide

## Quick Start with Podman

### 1. Build the Image

```powershell
cd headless_form
podman build -t headless-form:enterprise .
```

### 2. Run Container

```powershell
podman run -d --name hfm -p 8080:8080 -v headless-form-data:/data headless-form:enterprise
```

### 3. Access Application

Open http://localhost:8080 and create your first admin account.

---

## Environment Variables

| Variable         | Default          | Description               |
| ---------------- | ---------------- | ------------------------- |
| `PORT`           | 8080             | Server port               |
| `DATA_DIR`       | /data            | SQLite database directory |
| `JWT_SECRET`     | (auto-generated) | JWT signing key           |
| `TOKEN_DURATION` | 168h             | JWT expiry (7 days)       |
| `SMTP_HOST`      | -                | SMTP server host          |
| `SMTP_PORT`      | 587              | SMTP port                 |
| `SMTP_USERNAME`  | -                | SMTP username             |
| `SMTP_PASSWORD`  | -                | SMTP password             |
| `SMTP_FROM`      | -                | Sender email              |
| `SMTP_ENABLED`   | false            | Enable email sending      |

---

## Docker/Podman Commands

### View Logs

```powershell
podman logs -f hfm
```

### Stop Container

```powershell
podman stop hfm
```

### Start Container

```powershell
podman start hfm
```

### Remove Container

```powershell
podman rm -f hfm
```

### Rebuild and Restart

```powershell
podman rm -f hfm
podman build -t headless-form:enterprise .
podman run -d --name hfm -p 8080:8080 -v headless-form-data:/data headless-form:enterprise
```

---

## Production Recommendations

### 1. Configure JWT Secret

Set a strong, random `JWT_SECRET` environment variable:

```powershell
podman run -d --name hfm -p 8080:8080 \
  -e JWT_SECRET="your-secure-random-secret-here" \
  -v headless-form-data:/data \
  headless-form:enterprise
```

### 2. Enable SMTP for Password Reset

```powershell
podman run -d --name hfm -p 8080:8080 \
  -e SMTP_HOST="smtp.example.com" \
  -e SMTP_PORT="587" \
  -e SMTP_USERNAME="user@example.com" \
  -e SMTP_PASSWORD="password" \
  -e SMTP_FROM="noreply@example.com" \
  -e SMTP_ENABLED="true" \
  -v headless-form-data:/data \
  headless-form:enterprise
```

### 3. Reverse Proxy (Nginx)

```nginx
server {
    listen 443 ssl;
    server_name forms.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4. Backup Data

```powershell
# Copy data volume
podman cp hfm:/data/headless_form.db ./backup/
```

---

## Troubleshooting

### Container fails to start

Check logs: `podman logs hfm`

### Database locked error

Check write permissions on volume

### JWT token invalid

Verify `JWT_SECRET` is consistent between restarts
