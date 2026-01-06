package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"headless_form/internal/core/domain"
)

// SettingsRepository implements settings storage in SQLite
type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retrieves site settings (there's only one row with id='default')
func (r *SettingsRepository) Get(ctx context.Context) (*domain.SiteSettings, error) {
	settings := &domain.SiteSettings{
		ID:      "default",
		Version: "1.0.0",
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT site_name, site_url, smtp_host, smtp_port, smtp_user, smtp_password,
		       smtp_from, smtp_from_name, smtp_secure, updated_at, updated_by
		FROM site_settings WHERE id = 'default'
	`)

	var siteName, siteURL, smtpHost, smtpUser, smtpPass, smtpFrom, smtpFromName, updatedBy sql.NullString
	var smtpPort sql.NullInt32
	var smtpSecure sql.NullBool
	var updatedAt sql.NullTime

	err := row.Scan(&siteName, &siteURL, &smtpHost, &smtpPort, &smtpUser, &smtpPass,
		&smtpFrom, &smtpFromName, &smtpSecure, &updatedAt, &updatedBy)
	if err == sql.ErrNoRows {
		// Return defaults
		settings.SiteName = "Headless Forms"
		settings.SiteURL = "http://localhost:8080"
		settings.SMTPPort = 587
		settings.SMTPSecure = true
		return settings, nil
	}
	if err != nil {
		return nil, err
	}

	settings.SiteName = siteName.String
	settings.SiteURL = siteURL.String
	settings.SMTPHost = smtpHost.String
	settings.SMTPPort = int(smtpPort.Int32)
	settings.SMTPUser = smtpUser.String
	settings.SMTPPassword = smtpPass.String
	settings.SMTPFrom = smtpFrom.String
	settings.SMTPFromName = smtpFromName.String
	settings.SMTPSecure = smtpSecure.Bool
	settings.UpdatedAt = updatedAt.Time
	settings.UpdatedBy = updatedBy.String

	return settings, nil
}

// Save stores site settings (upsert)
func (r *SettingsRepository) Save(ctx context.Context, settings *domain.SiteSettings) error {
	settings.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO site_settings (id, site_name, site_url, smtp_host, smtp_port, smtp_user, smtp_password,
		                           smtp_from, smtp_from_name, smtp_secure, updated_at, updated_by)
		VALUES ('default', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			site_name = excluded.site_name,
			site_url = excluded.site_url,
			smtp_host = excluded.smtp_host,
			smtp_port = excluded.smtp_port,
			smtp_user = excluded.smtp_user,
			smtp_password = CASE WHEN excluded.smtp_password = '' THEN site_settings.smtp_password ELSE excluded.smtp_password END,
			smtp_from = excluded.smtp_from,
			smtp_from_name = excluded.smtp_from_name,
			smtp_secure = excluded.smtp_secure,
			updated_at = excluded.updated_at,
			updated_by = excluded.updated_by
	`, settings.SiteName, settings.SiteURL, settings.SMTPHost, settings.SMTPPort,
		settings.SMTPUser, settings.SMTPPassword, settings.SMTPFrom, settings.SMTPFromName,
		settings.SMTPSecure, settings.UpdatedAt, settings.UpdatedBy)

	return err
}

// Compile-time interface check
var _ interface {
	Get(ctx context.Context) (*domain.SiteSettings, error)
	Save(ctx context.Context, settings *domain.SiteSettings) error
} = (*SettingsRepository)(nil)

// Unused import placeholder
var _ = json.Marshal
