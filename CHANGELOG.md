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

## [1.1.0] - 2026-01-08

### Added

- 👤 **User Profile Page** - Users can now update their name, email, and password
- 🔒 **Profile API Endpoints**:
  - `PUT /api/v1/auth/profile` - Update own profile (name, email)
  - `PUT /api/v1/auth/password` - Change own password with current password verification
  - `PUT /api/v1/users/{id}` - Admin update any user (includes role changes)
- 📄 **Submission API Endpoint**:
  - `GET /api/v1/submissions/{id}` - Fetch single submission by ID
- ✅ **Complete CRUD Operations** for all entities
- 🚀 **GitHub Actions CI/CD** - Automated testing, building, and Docker publishing
- 📝 **Documentation** - Complete entity audit with ERD and API reference
- 🔧 **Environment Configuration** - `.env` file support with `godotenv` auto-loading

### Changed

- Enhanced submission handlers with ownership verification
- Fixed frontend User role type to match backend (`super_admin | admin | user`)
- Improved security: mark-as-read/unread/delete now verify form ownership
- Email validation with regex pattern and normalization

### Security

- Added form ownership checks to all submission modification endpoints
- Non-owners can no longer modify submissions they don't have access to
- Email format validation prevents invalid email addresses

### Documentation

- Added `docs/ENTITY_AUDIT.md` - Complete entity reference with ERD
- Added `docs/PHASE1_PLAN.md` - Implementation planning documentation

### DevOps

- Added `.github/workflows/ci.yml` - CI/CD pipeline
- Added `.golangci.yml` - Go linter configuration
