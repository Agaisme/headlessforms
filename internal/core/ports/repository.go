package ports

import (
	"context"
	"headless_form/internal/core/domain"
)

// Repository defines the contract for Data Storage.
// This allows us to swap SQLite for Postgres without changing Business Logic.
type Repository interface {
	// Transaction support
	Tx(ctx context.Context, fn func(Repository) error) error

	// Sub-Repositories
	Form() FormRepository
	Submission() SubmissionRepository
	Stats() StatsRepository
	User() UserRepository
	PasswordReset() PasswordResetRepository
	Settings() SettingsRepository
}

type FormRepository interface {
	Create(ctx context.Context, form *domain.Form) error
	Update(ctx context.Context, form *domain.Form) error
	GetByPublicID(ctx context.Context, publicID string) (*domain.Form, error)
	GetByID(ctx context.Context, id string) (*domain.Form, error)
	List(ctx context.Context) ([]*domain.Form, error)
	ListPaginated(ctx context.Context, limit, offset int) ([]*domain.Form, int, error)
	ListByOwnerPaginated(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Form, int, error)
	Delete(ctx context.Context, id string) error
	IncrementSubmissionCount(ctx context.Context, formID string) error
}

type SubmissionRepository interface {
	Create(ctx context.Context, submission *domain.Submission) error
	GetByID(ctx context.Context, id string) (*domain.Submission, error)
	GetByFormID(ctx context.Context, formID string) ([]*domain.Submission, error)
	GetByFormIDPaginated(ctx context.Context, formID string, limit, offset int) ([]*domain.Submission, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.SubmissionStatus) error
	Delete(ctx context.Context, id string) error
}

type StatsRepository interface {
	GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error)
	GetFormStats(ctx context.Context, formID string) (*domain.FormStats, error)
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*domain.User, error)
	Count(ctx context.Context) (int, error)
}

type PasswordResetRepository interface {
	Create(ctx context.Context, token *domain.PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, tokenID string) error
	DeleteExpired(ctx context.Context) error
}

type SettingsRepository interface {
	Get(ctx context.Context) (*domain.SiteSettings, error)
	Save(ctx context.Context, settings *domain.SiteSettings) error
}
