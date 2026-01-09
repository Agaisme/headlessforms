# Database Migrations

This directory contains SQL migration files for HeadlessForms.

## Naming Convention

```
[version]_[description].sql
```

- `001_initial_schema.sql` - Initial database schema
- `002_add_spam_score.sql` - Add spam score to submissions

## Running Migrations

Currently, HeadlessForms auto-creates tables on startup using embedded schema.
These migration files serve as:

1. **Documentation** - Clear history of schema changes
2. **Reference** - For manual migrations if needed
3. **Future use** - Will be used by migration runner

## Future: Automated Migrations

A migration runner will be added in a future version to:

- Track applied migrations in a `schema_migrations` table
- Apply pending migrations automatically on startup
- Support rollback operations

## Manual Application

To apply manually in SQLite:

```bash
sqlite3 data/headlessforms.db < migrations/001_initial_schema.sql
```
