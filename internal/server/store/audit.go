package store

import (
	"context"
	"database/sql"

	"github.com/boonkerz/roster/internal/server/model"
)

// InsertAudit schreibt einen Audit-Eintrag (fehlertolerant vom Aufrufer behandelt).
func (s *Store) InsertAudit(ctx context.Context, e model.AuditEntry) error {
	if e.ID == "" {
		e.ID = newID()
	}
	var uid any
	if e.UserID != "" {
		uid = e.UserID
	}
	_, err := s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO audit_log (id, ts, user_id, username, action, method, path, status, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		e.ID, e.TS.UTC(), uid, e.Username, e.Action, e.Method, e.Path, e.Status, e.IP)
	return err
}

// ListAudit liefert die jüngsten Audit-Einträge.
func (s *Store) ListAudit(ctx context.Context, limit int) ([]model.AuditEntry, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, ts, user_id, username, action, method, path, status, ip
		FROM audit_log ORDER BY ts DESC LIMIT ?`), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AuditEntry
	for rows.Next() {
		var e model.AuditEntry
		var uid sql.NullString
		if err := rows.Scan(&e.ID, &e.TS, &uid, &e.Username, &e.Action, &e.Method, &e.Path, &e.Status, &e.IP); err != nil {
			return nil, err
		}
		e.UserID = uid.String
		out = append(out, e)
	}
	return out, rows.Err()
}
