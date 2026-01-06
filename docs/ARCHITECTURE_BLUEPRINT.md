# Architecture Blueprint: Headless Form Manager (Enterprise)

## 1. System Overview

**Name**: Headless Form Manager
**Type**: Self-Hosted Backend-as-a-Service (BaaS)
**Deployment**: Single Static Binary (Monolithic)

The system is designed as a **Modular Monolith**. It looks like a single app from the outside, but is structured as distinct modules internally to allow for future splitting or DB switching.

## 2. High-Level Architecture

```mermaid
graph TD
    User[Client Browser] -->|HTTPS| Binary[Single Binary (Go)]

    subgraph "Single Binary Process"
        Router[HTTP Router (StdLib)]

        subgraph "Frontend Layer"
            StaticFS[Embedded SvelteKit Assets]
        end

        subgraph "Backend Layer"
            MW[Middleware (Auth/Logger/CORS)]
            API[REST API Handlers]
            Service[Core Business Logic]
            Repo[Repository Interface]
        end

        Router --> MW
        MW -->|/api/*| API
        MW -->|/*| StaticFS

        API --> Service
        Service --> Repo
    end

    subgraph "Data Layer (Pluggable)"
        SQLite[(SQLite File)]
        Postgres[(PostgreSQL DB)]
    end

    Repo -->|Driver: sqlite| SQLite
    Repo -->|Driver: postgres| Postgres
```

## 3. Technology Stack (Enterprise Quality)

### A. Backend: Go (Golang)

- **Version**: 1.22+
- **Router**: `net/http` (Standard Library).
- **Embed**: `embed` package to bundle the frontend.
- **Validation**: explicit struct validation (no magic tags).

### B. Frontend: SvelteKit

- **Mode**: SPA (Single Page App) via `adapter-static`.
- **Language**: TypeScript.
- **Styling**: TailwindCSS.

### C. Data Access: Repository Pattern (The "Laravel-like" Switch)

We enforce strict separation of concerns via Interfaces.

- **Location**: `internal/core/ports/repository.go`
- **Contract**:
  ```go
  type Repository interface {
      Form() FormRepository
      Submission() SubmissionRepository
      Tx(ctx context.Context, fn func(Repository) error) error
  }
  ```

## 4. API Standardization (JSend Style)

To ensure "No Mess", every JSON response MUST follow this structure. We will enforce this via a helper package `internal/api/response`.

**Success Response (200 OK)**

```json
{
  "status": "success",
  "data": {
    "id": "123",
    "name": "Contact Form"
  }
}
```

**Error Response (400/500)**

```json
{
  "status": "error",
  "message": "Invalid email address",
  "code": "INVALID_INPUT"
}
```

## 5. Directory Structure (Standard Go Layout)

```
/headless_form
  /cmd
    /server          # Main entry point (main.go)
  /internal
    /core            # DOMAIN LOGIC (Pure Go)
      /domain        # Structs (Form, Submission)
      /ports         # Interfaces (Repository, Service)
      /service       # Business Logic Implementation
    /adapter         # ADAPTERS (External World)
      /api           # HTTP Handlers (Standardized)
      /storage       # DB Implementations (SQLite/PG)
  /web               # SvelteKit App
  /docs              # Business & Architecture Docs
```
