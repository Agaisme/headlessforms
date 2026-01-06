package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"headless_form/internal/core/domain"
	"headless_form/internal/core/ports"

	_ "github.com/lib/pq" // Postgres driver
)

type Store struct {
	db *sql.DB
}

func New(connString string) (*Store, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS forms (
		id TEXT PRIMARY KEY,
		public_id TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		notify_emails JSONB NOT NULL,
		allowed_origins JSONB NOT NULL,
		redirect_url TEXT,
		submission_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS submissions (
		id TEXT PRIMARY KEY,
		form_id TEXT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
		status TEXT NOT NULL DEFAULT 'unread',
		data JSONB NOT NULL,
		meta JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT,
		role TEXT DEFAULT 'viewer',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_forms_public_id ON forms(public_id);
	CREATE INDEX IF NOT EXISTS idx_submissions_form_id ON submissions(form_id);
	CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	_, err := s.db.Exec(schema)
	return err
}

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

func (s *Store) Settings() ports.SettingsRepository {
	return nil // Not implemented for postgres yet
}

func (s *Store) Tx(ctx context.Context, fn func(ports.Repository) error) error {
	return fn(s)
}

func (s *Store) Close() error {
	return s.db.Close()
}

// FormRepository for Postgres
type FormRepository struct {
	db *sql.DB
}

func (r *FormRepository) Create(ctx context.Context, f *domain.Form) error {
	return nil // Implement as needed
}

func (r *FormRepository) Update(ctx context.Context, f *domain.Form) error {
	return nil
}

func (r *FormRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.Form, error) {
	return nil, nil
}

func (r *FormRepository) GetByID(ctx context.Context, id string) (*domain.Form, error) {
	return nil, nil
}

func (r *FormRepository) List(ctx context.Context) ([]*domain.Form, error) {
	return nil, nil
}

func (r *FormRepository) ListPaginated(ctx context.Context, limit, offset int) ([]*domain.Form, int, error) {
	return nil, 0, nil
}

func (r *FormRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *FormRepository) IncrementSubmissionCount(ctx context.Context, formID string) error {
	return nil
}

func (r *FormRepository) ListByOwnerPaginated(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Form, int, error) {
	return nil, 0, nil // Postgres not implemented - using SQLite
}

// SubmissionRepository for Postgres
type SubmissionRepository struct {
	db *sql.DB
}

func (r *SubmissionRepository) Create(ctx context.Context, s *domain.Submission) error {
	return nil
}

func (r *SubmissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	return nil, nil
}

func (r *SubmissionRepository) GetByFormID(ctx context.Context, formID string) ([]*domain.Submission, error) {
	return nil, nil
}

func (r *SubmissionRepository) GetByFormIDPaginated(ctx context.Context, formID string, limit, offset int) ([]*domain.Submission, int, error) {
	return nil, 0, nil
}

func (r *SubmissionRepository) UpdateStatus(ctx context.Context, id string, status domain.SubmissionStatus) error {
	return nil
}

func (r *SubmissionRepository) Delete(ctx context.Context, id string) error {
	return nil
}

// StatsRepository for Postgres
type StatsRepository struct {
	db *sql.DB
}

func (r *StatsRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	return &domain.DashboardStats{}, nil
}

func (r *StatsRepository) GetFormStats(ctx context.Context, formID string) (*domain.FormStats, error) {
	return &domain.FormStats{FormID: formID}, nil
}

// UserRepository for Postgres
type UserRepository struct {
	db *sql.DB
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	return nil, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	return 0, nil
}

func (s *Store) PasswordReset() ports.PasswordResetRepository {
	return &PasswordResetRepository{db: s.db}
}

// PasswordResetRepository for Postgres
type PasswordResetRepository struct {
	db *sql.DB
}

func (r *PasswordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	return nil
}

func (r *PasswordResetRepository) GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	return nil, nil
}

func (r *PasswordResetRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	return nil
}

func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {
	return nil
}
