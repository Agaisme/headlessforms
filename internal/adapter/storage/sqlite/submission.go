package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"headless_form/internal/core/domain"
)

type SubmissionRepository struct {
	db *sql.DB
}

func (r *SubmissionRepository) Create(ctx context.Context, s *domain.Submission) error {
	query := `INSERT INTO submissions (id, form_id, status, data, meta, created_at) VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.FormID, s.Status, string(s.Data), string(s.Meta), s.CreatedAt,
	)
	return err
}

func (r *SubmissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	query := `SELECT id, form_id, COALESCE(status, 'unread'), data, meta, created_at FROM submissions WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var s domain.Submission
	var dataRaw, metaRaw []byte

	if err := row.Scan(&s.ID, &s.FormID, &s.Status, &dataRaw, &metaRaw, &s.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan submission: %w", err)
	}

	s.Data = json.RawMessage(dataRaw)
	s.Meta = json.RawMessage(metaRaw)

	return &s, nil
}

func (r *SubmissionRepository) GetByFormID(ctx context.Context, formID string) ([]*domain.Submission, error) {
	query := `SELECT id, form_id, COALESCE(status, 'unread'), data, meta, created_at FROM submissions WHERE form_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, formID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var submissions []*domain.Submission
	for rows.Next() {
		var s domain.Submission
		var dataRaw, metaRaw []byte

		if err := rows.Scan(&s.ID, &s.FormID, &s.Status, &dataRaw, &metaRaw, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Data = json.RawMessage(dataRaw)
		s.Meta = json.RawMessage(metaRaw)
		submissions = append(submissions, &s)
	}
	return submissions, nil
}

func (r *SubmissionRepository) UpdateStatus(ctx context.Context, id string, status domain.SubmissionStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE submissions SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *SubmissionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM submissions WHERE id = ?`, id)
	return err
}

func (r *SubmissionRepository) GetByFormIDPaginated(ctx context.Context, formID string, limit, offset int) ([]*domain.Submission, int, error) {
	// Get total count
	var total int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE form_id = ?`, formID).Scan(&total)

	// Get paginated submissions
	query := `SELECT id, form_id, COALESCE(status, 'unread'), data, meta, created_at FROM submissions WHERE form_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, formID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var submissions []*domain.Submission
	for rows.Next() {
		var s domain.Submission
		var dataRaw, metaRaw []byte

		if err := rows.Scan(&s.ID, &s.FormID, &s.Status, &dataRaw, &metaRaw, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		s.Data = json.RawMessage(dataRaw)
		s.Meta = json.RawMessage(metaRaw)
		submissions = append(submissions, &s)
	}
	return submissions, total, nil
}
