package sqlite

import (
	"context"
	"database/sql"
	"time"

	"headless_form/internal/core/domain"
)

type PasswordResetRepository struct {
	db *sql.DB
}

func (r *PasswordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO password_resets (id, user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *PasswordResetRepository) GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_resets
		WHERE token = ?
	`, token)

	t := &domain.PasswordResetToken{}
	var usedAt sql.NullTime
	err := row.Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &usedAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if usedAt.Valid {
		t.UsedAt = &usedAt.Time
	}
	return t, nil
}

func (r *PasswordResetRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE password_resets SET used_at = ? WHERE id = ?
	`, time.Now(), tokenID)
	return err
}

func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM password_resets WHERE expires_at < datetime('now')
	`)
	return err
}
