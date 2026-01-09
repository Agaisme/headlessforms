package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"headless_form/internal/core/domain"
)

type FormRepository struct {
	db *sql.DB
}

func (r *FormRepository) Create(ctx context.Context, f *domain.Form) error {
	// Use INSERT with only the columns that always exist, plus new ones if available
	query := `INSERT INTO forms (id, public_id, name, notify_emails, allowed_origins, redirect_url, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	emailsJson, _ := json.Marshal(f.NotifyEmails)
	originsJson, _ := json.Marshal(f.AllowedOrigins)

	_, err := r.db.ExecContext(ctx, query,
		f.ID, f.PublicID, f.Name, string(emailsJson), string(originsJson),
		f.RedirectURL, f.CreatedAt,
	)

	// Try to set new columns - ignore errors if they don't exist
	if err == nil {
		_, _ = r.db.ExecContext(ctx, `UPDATE forms SET status = ?, submission_count = ?, updated_at = ?, webhook_url = ?, webhook_secret = ?, access_mode = ?, submission_key = ?, owner_id = ? WHERE id = ?`,
			f.Status, f.SubmissionCount, f.UpdatedAt, f.WebhookURL, f.WebhookSecret, f.AccessMode, f.SubmissionKey, f.OwnerID, f.ID)
	}

	return err
}

func (r *FormRepository) Update(ctx context.Context, f *domain.Form) error {
	query := `UPDATE forms SET name = ?, notify_emails = ?, allowed_origins = ?, redirect_url = ? WHERE id = ?`

	emailsJson, _ := json.Marshal(f.NotifyEmails)
	originsJson, _ := json.Marshal(f.AllowedOrigins)

	_, err := r.db.ExecContext(ctx, query,
		f.Name, string(emailsJson), string(originsJson), f.RedirectURL, f.ID,
	)

	// Try to set new columns - ignore errors if they don't exist
	if err == nil {
		_, _ = r.db.ExecContext(ctx, `UPDATE forms SET status = ?, updated_at = ?, webhook_url = ?, webhook_secret = ?, access_mode = ?, submission_key = ? WHERE id = ?`,
			f.Status, f.UpdatedAt, f.WebhookURL, f.WebhookSecret, f.AccessMode, f.SubmissionKey, f.ID)
	}

	return err
}

func (r *FormRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.Form, error) {
	return r.getByField(ctx, "public_id", publicID)
}

func (r *FormRepository) GetByID(ctx context.Context, id string) (*domain.Form, error) {
	return r.getByField(ctx, "id", id)
}

func (r *FormRepository) getByField(ctx context.Context, field, value string) (*domain.Form, error) {
	// G201: field is internal constant ("id" or "public_id"), not user input
	query := fmt.Sprintf(`SELECT id, public_id, name, notify_emails, allowed_origins, redirect_url, created_at FROM forms WHERE %s = ?`, field) // #nosec G201

	row := r.db.QueryRowContext(ctx, query, value)

	var f domain.Form
	var emailsRaw, originsRaw string

	if err := row.Scan(&f.ID, &f.PublicID, &f.Name, &emailsRaw, &originsRaw, &f.RedirectURL, &f.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan form: %w", err)
	}

	_ = json.Unmarshal([]byte(emailsRaw), &f.NotifyEmails)
	_ = json.Unmarshal([]byte(originsRaw), &f.AllowedOrigins)

	// Set defaults for new columns
	f.Status = domain.FormStatusActive
	f.SubmissionCount = 0
	f.UpdatedAt = f.CreatedAt

	// Try to read new columns if they exist
	var status sql.NullString
	var count int
	var webhookURL, webhookSecret, accessMode, submissionKey, ownerID sql.NullString
	// G201: field is internal constant, not user input
	extQuery := fmt.Sprintf(`SELECT status, submission_count, webhook_url, webhook_secret, access_mode, submission_key, owner_id FROM forms WHERE %s = ?`, field) // #nosec G201
	if err := r.db.QueryRowContext(ctx, extQuery, value).Scan(&status, &count, &webhookURL, &webhookSecret, &accessMode, &submissionKey, &ownerID); err == nil {
		if status.Valid && status.String != "" {
			f.Status = domain.FormStatus(status.String)
		}
		f.SubmissionCount = count
		f.WebhookURL = webhookURL.String
		f.WebhookSecret = webhookSecret.String
		if accessMode.Valid && accessMode.String != "" {
			f.AccessMode = accessMode.String
		} else {
			f.AccessMode = "public"
		}
		f.SubmissionKey = submissionKey.String
		f.OwnerID = ownerID.String
	}

	return &f, nil
}

func (r *FormRepository) List(ctx context.Context) ([]*domain.Form, error) {
	// Use only original columns for compatibility
	query := `SELECT id, public_id, name, notify_emails, allowed_origins, redirect_url, created_at FROM forms ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var forms []*domain.Form
	for rows.Next() {
		var f domain.Form
		var emailsRaw, originsRaw string
		if err := rows.Scan(&f.ID, &f.PublicID, &f.Name, &emailsRaw, &originsRaw, &f.RedirectURL, &f.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(emailsRaw), &f.NotifyEmails)
		_ = json.Unmarshal([]byte(originsRaw), &f.AllowedOrigins)

		// Set defaults
		f.Status = domain.FormStatusActive
		f.SubmissionCount = 0
		f.UpdatedAt = f.CreatedAt

		forms = append(forms, &f)
	}

	// Try to get extended data for all forms
	for _, f := range forms {
		var status string
		var count int
		if err := r.db.QueryRowContext(ctx, `SELECT status, submission_count FROM forms WHERE id = ?`, f.ID).Scan(&status, &count); err == nil {
			if status != "" {
				f.Status = domain.FormStatus(status)
			}
			f.SubmissionCount = count
		}
	}

	return forms, nil
}

func (r *FormRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM forms WHERE id = ?`, id)
	return err
}

func (r *FormRepository) ListPaginated(ctx context.Context, limit, offset int) ([]*domain.Form, int, error) {
	// Get total count
	var total int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM forms`).Scan(&total)

	// Get paginated forms
	query := `SELECT id, public_id, name, notify_emails, allowed_origins, redirect_url, created_at FROM forms ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var forms []*domain.Form
	for rows.Next() {
		var f domain.Form
		var emailsRaw, originsRaw string
		if err := rows.Scan(&f.ID, &f.PublicID, &f.Name, &emailsRaw, &originsRaw, &f.RedirectURL, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal([]byte(emailsRaw), &f.NotifyEmails)
		_ = json.Unmarshal([]byte(originsRaw), &f.AllowedOrigins)

		f.Status = domain.FormStatusActive
		f.SubmissionCount = 0
		f.UpdatedAt = f.CreatedAt

		forms = append(forms, &f)
	}

	// Try to get extended data for all forms
	for _, f := range forms {
		var status string
		var count int
		if err := r.db.QueryRowContext(ctx, `SELECT status, submission_count FROM forms WHERE id = ?`, f.ID).Scan(&status, &count); err == nil {
			if status != "" {
				f.Status = domain.FormStatus(status)
			}
			f.SubmissionCount = count
		}
	}

	return forms, total, nil
}

func (r *FormRepository) IncrementSubmissionCount(ctx context.Context, formID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE forms SET submission_count = COALESCE(submission_count, 0) + 1 WHERE id = ?`, formID)
	return err
}

func (r *FormRepository) ListByOwnerPaginated(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Form, int, error) {
	// Get total count for this owner
	var total int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM forms WHERE owner_id = ?`, ownerID).Scan(&total)

	// Get paginated forms for this owner
	query := `SELECT id, public_id, name, notify_emails, allowed_origins, redirect_url, created_at FROM forms WHERE owner_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var forms []*domain.Form
	for rows.Next() {
		var f domain.Form
		var emailsRaw, originsRaw string
		if err := rows.Scan(&f.ID, &f.PublicID, &f.Name, &emailsRaw, &originsRaw, &f.RedirectURL, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal([]byte(emailsRaw), &f.NotifyEmails)
		_ = json.Unmarshal([]byte(originsRaw), &f.AllowedOrigins)

		f.Status = domain.FormStatusActive
		f.SubmissionCount = 0
		f.UpdatedAt = f.CreatedAt
		f.OwnerID = ownerID

		forms = append(forms, &f)
	}

	// Try to get extended data for all forms
	for _, f := range forms {
		var status string
		var count int
		if err := r.db.QueryRowContext(ctx, `SELECT status, submission_count FROM forms WHERE id = ?`, f.ID).Scan(&status, &count); err == nil {
			if status != "" {
				f.Status = domain.FormStatus(status)
			}
			f.SubmissionCount = count
		}
	}

	return forms, total, nil
}
