# Business Process Documentation

## 1. Core Workflows

### A. Creating a New Form (Admin)

1.  **Actor**: Admin User (Logged in).
2.  **Action**: Clicks "Create Form" -> Enters "Name", "Redirect URL", "Notify Email".
3.  **System**:
    - Generates a unique `PublicID` (UUID).
    - Creates a `SecretKey` (for secure server-side posting) if needed.
    - Saves to DB.
4.  **Result**: Admin receives a Snippet (HTML/JS) to copy-paste.

### B. Submitting Data (Public User)

1.  **Actor**: Website Visitor.
2.  **Action**: Fills out a form on `client-website.com`.
3.  **System (Frontend SDK)**:
    - Captures form data.
    - POSTs to `https://api.headless-form.com/v1/submit/{PublicID}`.
4.  **System (Backend)**:
    - Validates Origin (CORS info).
    - Validates Payload (Basic Type Checks).
    - Saves Submission to DB.
    - **Async**: Queues Email Notification.
5.  **Result**: Returns 200 OK. Frontend redirects user to "Thank You" page.

### C. Email Notification

1.  **Trigger**: New valid submission.
2.  **Logic**: Check `Form.NotifyEmail`.
3.  **Action**: Render Email Template (Go Templates) -> Send via SMTP.

## 2. API Contract (Draft)

### Public Endpoints

- `POST /v1/submissions/{form_id}`
  - Body: `{"data": {...}, "meta": {...}}`
  - Response: `200 OK`

### Admin Endpoints (Protected)

- `GET /v1/forms` - List all forms.
- `POST /v1/forms` - Create form.
- `GET /v1/forms/{id}/submissions` - View data.
- `POST /v1/auth/login` - Admin login.

## 3. Security Requirements

- **Rate Limiting**: Prevent spam attacks per IP.
- **Honeypot**: Hidden fields to catch bots.
- **Data Sanitation**: Prevent XSS in the Admin Dashboard.
