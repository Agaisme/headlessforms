package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"headless_form/internal/core/ports"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable WAL mode for concurrency
	if _, err := db.Exec(`PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	// Base schema - compatible with existing databases
	schema := `
	CREATE TABLE IF NOT EXISTS forms (
		id TEXT PRIMARY KEY,
		public_id TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		notify_emails TEXT NOT NULL,
		allowed_origins TEXT NOT NULL,
		redirect_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS submissions (
		id TEXT PRIMARY KEY,
		form_id TEXT NOT NULL,
		data JSON NOT NULL,
		meta JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(form_id) REFERENCES forms(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_forms_public_id ON forms(public_id);
	CREATE INDEX IF NOT EXISTS idx_submissions_form_id ON submissions(form_id);
	`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}

	// Users table
	usersSchema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT,
		role TEXT DEFAULT 'viewer',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	_, _ = s.db.Exec(usersSchema)

	// Run migrations for new columns (ignore errors if columns already exist)
	migrations := []string{
		`ALTER TABLE forms ADD COLUMN status TEXT DEFAULT 'active'`,
		`ALTER TABLE forms ADD COLUMN submission_count INTEGER DEFAULT 0`,
		`ALTER TABLE forms ADD COLUMN updated_at DATETIME`,
		`ALTER TABLE forms ADD COLUMN webhook_url TEXT`,
		`ALTER TABLE forms ADD COLUMN webhook_secret TEXT`,
		`ALTER TABLE forms ADD COLUMN access_mode TEXT DEFAULT 'public'`,
		`ALTER TABLE forms ADD COLUMN submission_key TEXT`,
		`ALTER TABLE forms ADD COLUMN owner_id TEXT`,
		`ALTER TABLE submissions ADD COLUMN status TEXT DEFAULT 'unread'`,
	}

	for _, m := range migrations {
		_, _ = s.db.Exec(m)
	}

	// Create indexes for new columns
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_created_at ON submissions(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_forms_owner_id ON forms(owner_id)`,
	}

	for _, idx := range indexes {
		_, _ = s.db.Exec(idx)
	}

	// Reset tokens table
	//nolint:gosec // G101 false positive - this is a table schema, not credentials
	resetTokensSchema := `
	CREATE TABLE IF NOT EXISTS password_resets (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		used_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
	CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
	`
	_, _ = s.db.Exec(resetTokensSchema)

	// Site settings table
	siteSettingsSchema := `
	CREATE TABLE IF NOT EXISTS site_settings (
		id TEXT PRIMARY KEY DEFAULT 'default',
		site_name TEXT,
		site_url TEXT,
		smtp_host TEXT,
		smtp_port INTEGER DEFAULT 587,
		smtp_user TEXT,
		smtp_password TEXT,
		smtp_from TEXT,
		smtp_from_name TEXT,
		smtp_secure INTEGER DEFAULT 1,
		updated_at DATETIME,
		updated_by TEXT
	);
	`
	_, _ = s.db.Exec(siteSettingsSchema)

	return nil
}

// Implement Repository Interface
func (s *Store) Form() ports.FormRepository {
	return &FormRepository{db: s.db}
}

func (s *Store) Submission() ports.SubmissionRepository {
	return &SubmissionRepository{db: s.db}
}

func (s *Store) Stats() ports.StatsRepository {
	return &StatsRepository{db: s.db}
}

func (s *Store) User() ports.UserRepository {
	return &UserRepository{db: s.db}
}

func (s *Store) PasswordReset() ports.PasswordResetRepository {
	return &PasswordResetRepository{db: s.db}
}

func (s *Store) Settings() ports.SettingsRepository {
	return &SettingsRepository{db: s.db}
}

func (s *Store) Tx(ctx context.Context, fn func(ports.Repository) error) error {
	return fn(s)
}

func (s *Store) Close() error {
	return s.db.Close()
}
