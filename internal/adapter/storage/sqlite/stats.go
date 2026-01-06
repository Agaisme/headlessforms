package sqlite

import (
	"context"
	"database/sql"
	"headless_form/internal/core/domain"
)

type StatsRepository struct {
	db *sql.DB
}

func (r *StatsRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{}

	// Total forms
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM forms`).Scan(&stats.TotalForms)

	// Active forms
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM forms WHERE status = 'active' OR status IS NULL`).Scan(&stats.ActiveForms)

	// Total submissions
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions`).Scan(&stats.TotalSubmissions)

	// Unread submissions
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE status = 'unread' OR status IS NULL`).Scan(&stats.UnreadSubmissions)

	// Submissions today
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE date(created_at) = date('now')`).Scan(&stats.SubmissionsToday)

	// Submissions this week
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE created_at >= date('now', '-7 days')`).Scan(&stats.SubmissionsThisWeek)

	// Daily submissions for the last 7 days (for chart)
	rows, err := r.db.QueryContext(ctx, `
		WITH RECURSIVE dates(date) AS (
			SELECT date('now', '-6 days')
			UNION ALL
			SELECT date(date, '+1 day')
			FROM dates
			WHERE date < date('now')
		)
		SELECT d.date, COALESCE(COUNT(s.id), 0) as count
		FROM dates d
		LEFT JOIN submissions s ON date(s.created_at) = d.date
		GROUP BY d.date
		ORDER BY d.date
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var daily domain.DailySubmission
			if err := rows.Scan(&daily.Date, &daily.Count); err == nil {
				stats.DailySubmissions = append(stats.DailySubmissions, daily)
			}
		}
	}

	return stats, nil
}

func (r *StatsRepository) GetFormStats(ctx context.Context, formID string) (*domain.FormStats, error) {
	stats := &domain.FormStats{FormID: formID}

	// Total submissions for this form
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE form_id = ?`, formID).Scan(&stats.TotalSubmissions)

	// Unread submissions
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE form_id = ? AND (status = 'unread' OR status IS NULL)`, formID).Scan(&stats.UnreadSubmissions)

	// Submissions today
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE form_id = ? AND date(created_at) = date('now')`, formID).Scan(&stats.SubmissionsToday)

	// Submissions this week
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE form_id = ? AND created_at >= date('now', '-7 days')`, formID).Scan(&stats.SubmissionsThisWeek)

	return stats, nil
}
