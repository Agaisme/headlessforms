# API Versioning Strategy

## Current Version

**v1** - Stable, production-ready

All endpoints are prefixed with `/api/v1/`.

## Versioning Approach

HeadlessForms uses **URI versioning** for clarity and simplicity:

```
/api/v1/forms
/api/v1/submissions/{id}
/api/v2/forms  (future)
```

### Why URI Versioning?

| Method      | Pros             | Cons               |
| ----------- | ---------------- | ------------------ |
| URI (ours)  | Clear, cacheable | Multiple codepaths |
| Header      | Clean URLs       | Less visible       |
| Query param | Easy to test     | Not RESTful        |

## Breaking vs Non-Breaking Changes

### Non-Breaking (No version bump)

- Adding new optional fields
- Adding new endpoints
- Adding new optional query parameters
- Deprecation announcements

### Breaking (Version bump required)

- Removing fields
- Changing field types
- Changing response structure
- Removing endpoints

## Deprecation Policy

1. **Announce** - Mark deprecated in docs and headers
2. **Warn** - Add `Deprecation` header to responses
3. **Sunset** - Remove after 6 months minimum

### Deprecation Header Example

```http
Deprecation: true
Sunset: Sat, 01 Jul 2027 00:00:00 GMT
Link: </api/v2/forms>; rel="successor-version"
```

## Changelog

### v1.0.0 (2026-01-07)

- Initial release
- Full CRUD for forms and submissions
- JWT authentication

### v1.1.0 (2026-01-08)

- Added profile endpoints
- Added password reset flow

### v1.2.0 (2026-01-10)

- Enhanced health check
- Integration tests
- No breaking changes
