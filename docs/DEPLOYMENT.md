# Deployment Guide

## 1. Single Binary Deployment (Recommended)

The easiest way to deploy is as a single executable that contains both the backend and frontend.

### Prerequisites

- Go 1.22+
- Node.js 20+

### Build Steps

1. **Build Frontend**:

   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

   This generates static assets in `web/build`.

2. **Build Backend**:

   ```bash
   # Linux
   GOOS=linux GOARCH=amd64 go build -o headless-form ./cmd/server

   # Windows
   go build -o headless-form.exe ./cmd/server
   ```

3. **Deploy**:
   Upload the binary and your `.env` file to your server.

### Running

```bash
./headless-form
```

Access at `http://your-server:8080`.

---

## 2. Docker Deployment

### Build Image

```bash
docker build -t headless-form .
```

### Run Container

```bash
docker run -d \
  -p 8080:8080 \
  --env-file .env \
  --name headless-form \
  headless-form
```

---

## 3. Configuration (.env)

| Variable       | Description         | Example                                  |
| -------------- | ------------------- | ---------------------------------------- |
| `PORT`         | Server port         | `8080`                                   |
| `ENV`          | Environment         | `production`                             |
| `DATABASE_URL` | Postgres Connection | `postgres://user:pass@localhost:5432/db` |
| `JWT_SECRET`   | Secret Key          | `change-me-in-prod`                      |

See `.env.example` for full list.
