# Changelog

All notable changes to HeadlessForms will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-07

### Added

- 🚀 Initial release
- Single-binary deployment with embedded frontend
- SQLite database (default) with PostgreSQL support
- JWT-based authentication with role-based access control
- Form management with access modes (public, with_key, private)
- Email notifications via SMTP
- Webhook integration with HMAC-SHA256 signing
- Modern dashboard with analytics
- Mobile-responsive inbox view
- CSV export for submissions
- Rate limiting on authentication endpoints
- Multi-stage Docker build

### Security

- bcrypt password hashing
- JWT token authentication
- CORS configuration
- SQL injection protection (parameterized queries)
- Role-based access control (user, admin, super_admin)

## [Unreleased]

### Planned

- Dark mode toggle
- Keyboard shortcuts for inbox navigation
- Bulk submission actions
- Form templates
- API rate limiting per form
