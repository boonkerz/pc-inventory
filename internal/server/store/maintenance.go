package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/thomaspeterson/pc-inventory/internal/server/model"
)

// CreateMaintenanceWindow legt ein Wartungsfenster an.
func (s *Store) CreateMaintenanceWindow(ctx context.Context, m *model.MaintenanceWindow) error {
	m.ID = newID()
	m.CreatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO maintenance_windows (id, target_type, target_id, note, starts_at, ends_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`),
		m.ID, m.TargetType, m.TargetID, m.Note, m.StartsAt.UTC(), m.EndsAt.UTC(), m.CreatedAt)
	return err
}

// DeleteMaintenanceWindow entfernt ein Wartungsfenster.
func (s *Store) DeleteMaintenanceWindow(ctx context.Context, id string) error {
	return s.affect(s.db.ExecContext(ctx, s.rebind(`DELETE FROM maintenance_windows WHERE id=?`), id))
}

// ListMaintenanceWindows liefert alle Fenster (jüngste zuerst) inkl. Zielnamen.
// Abgelaufene Fenster werden dabei aufgeräumt.
func (s *Store) ListMaintenanceWindows(ctx context.Context) ([]model.MaintenanceWindow, error) {
	_, _ = s.db.ExecContext(ctx, s.rebind(`DELETE FROM maintenance_windows WHERE ends_at < ?`),
		time.Now().Add(-24*time.Hour).UTC())
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT m.id, m.target_type, m.target_id, m.note, m.starts_at, m.ends_at, m.created_at,
			COALESCE(c.name, s.name, d.hostname, m.target_id)
		FROM maintenance_windows m
		LEFT JOIN clients c ON m.target_type='client' AND c.id=m.target_id
		LEFT JOIN sites   s ON m.target_type='site'   AND s.id=m.target_id
		LEFT JOIN devices d ON m.target_type='device' AND d.id=m.target_id
		ORDER BY m.starts_at DESC`))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.MaintenanceWindow
	for rows.Next() {
		var m model.MaintenanceWindow
		var name sql.NullString
		if err := rows.Scan(&m.ID, &m.TargetType, &m.TargetID, &m.Note, &m.StartsAt, &m.EndsAt, &m.CreatedAt, &name); err != nil {
			return nil, err
		}
		m.TargetName = name.String
		out = append(out, m)
	}
	return out, rows.Err()
}

// DeviceInMaintenance meldet, ob für ein Gerät gerade ein Wartungsfenster aktiv ist
// (direkt oder über dessen Site/Client).
func (s *Store) DeviceInMaintenance(ctx context.Context, deviceID string, now time.Time) (bool, error) {
	var siteID, clientID sql.NullString
	if err := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT d.site_id, s.client_id FROM devices d LEFT JOIN sites s ON s.id=d.site_id WHERE d.id=?`),
		deviceID).Scan(&siteID, &clientID); err != nil {
		return false, err
	}
	var n int
	err := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT COUNT(*) FROM maintenance_windows
		WHERE starts_at <= ? AND ends_at >= ? AND (
			(target_type='device' AND target_id=?) OR
			(target_type='site'   AND target_id=?) OR
			(target_type='client' AND target_id=?))`),
		now.UTC(), now.UTC(), deviceID, siteID.String, clientID.String).Scan(&n)
	return n > 0, err
}
