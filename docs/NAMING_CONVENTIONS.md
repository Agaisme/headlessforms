# Naming Conventions

## Backend (Go)

### Files

| Type     | Convention     | Example            |
| -------- | -------------- | ------------------ |
| Handlers | `handler_*.go` | `handler_forms.go` |
| Services | `service.go`   | `service.go`       |
| Models   | `model.go`     | `model.go`         |
| Tests    | `*_test.go`    | `handler_test.go`  |

### Functions

- PascalCase for exported: `HandleCreateForm`
- camelCase for internal: `parseIntParam`

### Variables

- camelCase: `formService`, `submissionCount`
- Constants: UPPER_SNAKE: `MaxPageSize`

---

## Frontend (Svelte/TypeScript)

### Files

| Type       | Convention         | Example            |
| ---------- | ------------------ | ------------------ |
| Components | PascalCase         | `FormModal.svelte` |
| Routes     | SvelteKit standard | `+page.svelte`     |
| Stores     | camelCase          | `auth.ts`          |
| Utils      | camelCase          | `utils.ts`         |
| Types      | camelCase          | `types.ts`         |

### Variables

- camelCase: `isLoading`, `formData`
- Constants: UPPER_SNAKE: `API_BASE_URL`

### Components

- PascalCase: `<Button>`, `<FormModal>`
- Props: camelCase: `isOpen`, `formId`

---

## Database

### Tables

- snake_case plural: `forms`, `submissions`, `password_resets`

### Columns

- snake_case: `created_at`, `owner_id`, `is_read`

### Indexes

- `idx_[table]_[column]`: `idx_forms_public_id`

---

## API

### Endpoints

- Plural nouns: `/forms`, `/submissions`
- Nested resources: `/forms/{id}/submissions`
- Actions: `/submissions/{id}/read`

### Query Parameters

- snake_case: `page`, `limit`, `form_id`

### JSON Fields

- snake_case: `form_id`, `created_at`, `is_read`
