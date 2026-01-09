# HeadlessForms

**A self-hosted, single-binary form backend for developers.**

Accept form submissions from any website. No frontend required. Just point your HTML forms at HeadlessForms and collect submissions with email notifications, webhooks, and a beautiful admin dashboard.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## ✨ Features

- **Single Binary** - One executable, zero dependencies
- **Embedded Database** - SQLite by default, PostgreSQL optional
- **Beautiful Dashboard** - Modern UI built with SvelteKit
- **Email Notifications** - Get notified on new submissions
- **Webhooks** - POST data to any URL with HMAC signing
- **Access Control** - Public, key-protected, or private forms
- **Multi-User** - Role-based access (Admin, User, Super Admin)
- **API First** - RESTful JSON API for everything
- **Docker Ready** - Deploy anywhere in seconds

---

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
docker run -d \
  --name headlessforms \
  -p 8080:8080 \
  -v ./data:/data \
  headlessforms/headlessforms:latest
```

Then open http://localhost:8080 and create your admin account.

### Option 2: Download Binary

1. Download from [Releases](https://github.com/yourorg/headlessforms/releases)
2. Run: `./headlessforms`
3. Open http://localhost:8080

### Option 3: Build from Source

```bash
git clone https://github.com/yourorg/headlessforms.git
cd headlessforms
go build -o headlessforms ./cmd/server
./headlessforms
```

---

## 📝 Usage

### 1. Create a Form

Log into the dashboard and click **New Form**.

### 2. Point Your HTML Form

```html
<form action="https://your-server.com/api/v1/submissions/FORM_ID" method="POST">
  <input type="text" name="name" placeholder="Your Name" required />
  <input type="email" name="email" placeholder="Email" required />
  <textarea name="message" placeholder="Message"></textarea>
  <button type="submit">Send</button>
</form>
```

### 3. View Submissions

Go to **Forms** in the dashboard to see all submissions in a beautiful inbox view.

---

## ⚙️ Configuration

HeadlessForms uses environment variables for configuration:

| Variable     | Default        | Description                              |
| ------------ | -------------- | ---------------------------------------- |
| `PORT`       | `8080`         | Server port                              |
| `DATA_DIR`   | `./data`       | Database storage directory               |
| `JWT_SECRET` | Auto-generated | JWT signing secret (set for production!) |

### Docker Example

```bash
docker run -d \
  -p 8080:8080 \
  -v ./data:/data \
  -e JWT_SECRET="your-super-secret-key" \
  headlessforms/headlessforms:latest
```

---

## 🔒 Access Modes

| Mode         | Description                  | Use Case                   |
| ------------ | ---------------------------- | -------------------------- |
| **Public**   | Anyone can submit            | Contact forms, newsletters |
| **With Key** | Requires hidden `_key` field | Spam protection            |
| **Private**  | Requires authentication      | Internal forms             |

### Using Key Protection

1. Set form to "With Key" mode
2. Generate a submission key
3. Add hidden field to your form:

```html
<input type="hidden" name="_submission_key" value="YOUR_SUBMISSION_KEY" />
```

---

## 📧 Email Notifications

Configure SMTP in **Settings** to receive email notifications:

1. Go to Settings → Email Configuration
2. Enter your SMTP server details
3. Add notification emails to your forms

---

## 🪝 Webhooks

HeadlessForms can POST submissions to any URL:

1. Edit your form
2. Add a Webhook URL
3. Optionally add a Webhook Secret for HMAC-SHA256 signing

The webhook payload includes:

```json
{
  "form_id": "...",
  "submission_id": "...",
  "data": { ... },
  "created_at": "..."
}
```

---

## 🛠️ API Reference

All endpoints return JSend-style JSON responses. See [OpenAPI Spec](./docs/openapi.yaml) for full documentation.

### Endpoints Overview

| Method   | Endpoint                         | Auth   | Description                               |
| -------- | -------------------------------- | ------ | ----------------------------------------- |
| `POST`   | `/api/v1/auth/login`             | No     | Login, get JWT token                      |
| `POST`   | `/api/v1/auth/register`          | No     | Register (first user becomes super_admin) |
| `GET`    | `/api/v1/auth/me`                | Yes    | Get current user info                     |
| `GET`    | `/api/v1/forms`                  | Yes    | List forms (paginated)                    |
| `POST`   | `/api/v1/forms`                  | Yes    | Create new form                           |
| `GET`    | `/api/v1/forms/{id}`             | Yes    | Get form details                          |
| `PUT`    | `/api/v1/forms/{id}`             | Yes    | Update form                               |
| `DELETE` | `/api/v1/forms/{id}`             | Yes    | Delete form                               |
| `GET`    | `/api/v1/forms/{id}/submissions` | Yes    | List submissions                          |
| `GET`    | `/api/v1/forms/{id}/export/csv`  | Yes    | Export as CSV                             |
| `POST`   | `/api/v1/submissions/{id}`       | Varies | Submit to form                            |
| `PUT`    | `/api/v1/submissions/{id}/read`  | Yes    | Mark as read                              |
| `DELETE` | `/api/v1/submissions/{id}`       | Yes    | Delete submission                         |
| `GET`    | `/api/v1/stats`                  | Yes    | Dashboard statistics                      |
| `GET`    | `/api/v1/users`                  | Admin  | List users                                |
| `POST`   | `/api/v1/users`                  | Admin  | Create user                               |
| `GET`    | `/api/v1/settings`               | Super  | Get settings                              |
| `PUT`    | `/api/v1/settings`               | Super  | Update settings                           |

### Example: Create Form

```bash
curl -X POST http://localhost:8080/api/v1/forms \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Contact Form", "notify_emails": ["you@example.com"]}'
```

### Example: Submit to Form

```bash
curl -X POST http://localhost:8080/api/v1/submissions/FORM_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com", "message": "Hello!"}'
```

See [API Documentation](./docs/API.md) for complete reference.

---

## 📁 Project Structure

```
headlessforms/
├── cmd/server/         # Application entry point
├── internal/
│   ├── adapter/api/    # HTTP handlers
│   ├── adapter/storage # Database implementations
│   ├── core/domain/    # Business entities
│   ├── core/service/   # Business logic
│   └── core/ports/     # Interface definitions
├── web/                # SvelteKit frontend (embedded)
└── Dockerfile
```

---

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**Made with ❤️ for developers who just want forms to work.**
