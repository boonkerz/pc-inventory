package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/boonkerz/roster/internal/server/model"
)

// CreateReportSchedule legt einen geplanten Bericht an.
func (s *Store) CreateReportSchedule(ctx context.Context, r *model.ReportSchedule) error {
	r.ID = newID()
	r.CreatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO report_schedules (id, title, frequency, channel_id, last_run, created_at)
		VALUES (?, ?, ?, ?, NULL, ?)`),
		r.ID, r.Title, r.Frequency, r.ChannelID, r.CreatedAt)
	return err
}

// DeleteReportSchedule entfernt einen geplanten Bericht.
func (s *Store) DeleteReportSchedule(ctx context.Context, id string) error {
	return s.affect(s.db.ExecContext(ctx, s.rebind(`DELETE FROM report_schedules WHERE id=?`), id))
}

// ListReportSchedules liefert alle geplanten Berichte inkl. Kanalname.
func (s *Store) ListReportSchedules(ctx context.Context) ([]model.ReportSchedule, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT rs.id, rs.title, rs.frequency, rs.channel_id, rs.last_run, rs.created_at, COALESCE(c.name, '')
		FROM report_schedules rs LEFT JOIN alert_channels c ON c.id = rs.channel_id
		ORDER BY rs.created_at DESC`))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReportSchedules(rows)
}

// DueReportSchedules liefert Berichte, die laut Häufigkeit fällig sind (last_run alt genug).
func (s *Store) DueReportSchedules(ctx context.Context, now time.Time) ([]model.ReportSchedule, error) {
	all, err := s.ListReportSchedules(ctx)
	if err != nil {
		return nil, err
	}
	var due []model.ReportSchedule
	for _, r := range all {
		if reportDue(r.Frequency, r.LastRun, now) {
			due = append(due, r)
		}
	}
	return due, nil
}

// MarkReportRun setzt den Zeitpunkt des letzten Versands.
func (s *Store) MarkReportRun(ctx context.Context, id string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, s.rebind(`UPDATE report_schedules SET last_run=? WHERE id=?`), at.UTC(), id)
	return err
}

func reportDue(freq string, last *time.Time, now time.Time) bool {
	if last == nil {
		return true
	}
	elapsed := now.Sub(*last)
	switch freq {
	case "weekly":
		return elapsed >= 7*24*time.Hour
	case "monthly":
		return elapsed >= 30*24*time.Hour
	default: // daily
		return elapsed >= 24*time.Hour
	}
}

func scanReportSchedules(rows *sql.Rows) ([]model.ReportSchedule, error) {
	var out []model.ReportSchedule
	for rows.Next() {
		var r model.ReportSchedule
		var last sql.NullTime
		if err := rows.Scan(&r.ID, &r.Title, &r.Frequency, &r.ChannelID, &last, &r.CreatedAt, &r.ChannelName); err != nil {
			return nil, err
		}
		if last.Valid {
			t := last.Time
			r.LastRun = &t
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
