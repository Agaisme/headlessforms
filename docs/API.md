# HeadlessForms API Reference

## Base URL

```
http://localhost:8080/api/v1
```

---

## Authentication

### Register User

`POST /auth/register`

```json
{ "email": "user@example.com", "password": "securepass", "name": "John" }
```

### Login

`POST /auth/login`

```json
{ "email": "user@example.com", "password": "securepass" }
```

**Response:** `{ "token": "jwt...", "user": {...} }`

### Get Current User

`GET /auth/me`  
**Header:** `Authorization: Bearer {token}`

### Forgot Password

`POST /auth/forgot-password`

```json
{ "email": "user@example.com" }
```

### Reset Password

`POST /auth/reset-password`

```json
{ "token": "reset_token", "new_password": "newpassword123" }
```

---

## Forms

### Create Form

`POST /forms`

```json
{ "name": "Contact Form", "redirect_url": "https://example.com/thank-you" }
```

### List Forms

`GET /forms?page=1&limit=20`

### Get Form

`GET /forms/{form_id}`

### Update Form

`PUT /forms/{form_id}`

```json
{
  "name": "Updated Name",
  "status": "active",
  "access_mode": "public",
  "submission_key": "optional_key",
  "webhook_url": "https://hooks.example.com/webhook"
}
```

### Delete Form

`DELETE /forms/{form_id}`

---

## Submissions

### Submit Form (Public)

`POST /submissions/{form_id}`

```json
{ "email": "visitor@example.com", "message": "Hello!" }
```

**Access modes:**

- `public` - Anyone can submit
- `with_key` - Requires `_submission_key` field
- `private` - Requires JWT authentication

### List Submissions

`GET /forms/{form_id}/submissions?page=1&limit=20`

### Export CSV

`GET /forms/{form_id}/export/csv`  
**Returns:** CSV file download

### Mark as Read

`PUT /submissions/{sub_id}/read`

### Mark as Unread

`PUT /submissions/{sub_id}/unread`

### Delete Submission

`DELETE /submissions/{sub_id}`

---

## User Management (Admin)

### List Users

`GET /users`

### Create User

`POST /users`

```json
{
  "email": "new@example.com",
  "password": "pass123",
  "name": "New User",
  "role": "viewer"
}
```

### Delete User

`DELETE /users/{user_id}`

---

## Stats

### Dashboard Stats

`GET /stats`  
**Response:**

```json
{
  "total_forms": 10,
  "total_submissions": 245,
  "unread_submissions": 12,
  "submissions_this_week": 34,
  "daily_submissions": [{"date": "2026-01-01", "count": 5}, ...]
}
```

### Form Stats

`GET /forms/{form_id}/stats`

---

## Response Format

All responses follow JSend format:

```json
{
  "status": "success|fail|error",
  "data": { ... },
  "message": "Optional error message"
}
```
