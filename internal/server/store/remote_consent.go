package store

import (
	"context"
	"database/sql"
)

// Zustimmungs-Modus der Fernsteuerung: "unattended" (unbeaufsichtigt) oder "prompt"
// (der angemeldete Nutzer muss am Gerät bestätigen). Zielgruppenabhängig mit
// Vererbung device -> site -> client; ohne Eintrag gilt der Default "unattended".

const defaultRemoteConsent = "unattended"

// SetRemoteConsent setzt den Modus für ein Ziel; mode="" entfernt den Eintrag
// (dann greift die Vererbung/Default).
func (s *Store) SetRemoteConsent(ctx context.Context, targetType, targetID, mode string) error {
	if mode == "" {
		_, err := s.db.ExecContext(ctx, s.rebind(
			`DELETE FROM remote_consent WHERE target_type=? AND target_id=?`), targetType, targetID)
		return err
	}
	res, err := s.db.ExecContext(ctx, s.rebind(
		`UPDATE remote_consent SET mode=? WHERE target_type=? AND target_id=?`), mode, targetType, targetID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, s.rebind(
			`INSERT INTO remote_consent (target_type, target_id, mode) VALUES (?, ?, ?)`), targetType, targetID, mode)
	}
	return err
}

// GetRemoteConsent liefert den explizit gesetzten Modus eines Ziels ("" = keiner).
func (s *Store) GetRemoteConsent(ctx context.Context, targetType, targetID string) (string, error) {
	var mode string
	err := s.db.QueryRowContext(ctx, s.rebind(
		`SELECT mode FROM remote_consent WHERE target_type=? AND target_id=?`), targetType, targetID).Scan(&mode)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return mode, err
}

// ResolveRemoteConsent ermittelt den effektiven Modus für ein Gerät
// (device > site > client > Default).
func (s *Store) ResolveRemoteConsent(ctx context.Context, deviceID string) (string, error) {
	var siteID, clientID sql.NullString
	if err := s.db.QueryRowContext(ctx, s.rebind(
		`SELECT d.site_id, s.client_id FROM devices d LEFT JOIN sites s ON s.id=d.site_id WHERE d.id=?`),
		deviceID).Scan(&siteID, &clientID); err != nil {
		return defaultRemoteConsent, err
	}
	var mode string
	err := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT mode FROM remote_consent WHERE
			(target_type='device' AND target_id=?) OR
			(target_type='site'   AND target_id=?) OR
			(target_type='client' AND target_id=?)
		ORDER BY CASE target_type WHEN 'device' THEN 0 WHEN 'site' THEN 1 ELSE 2 END
		LIMIT 1`),
		deviceID, siteID.String, clientID.String).Scan(&mode)
	if err == sql.ErrNoRows || mode == "" {
		return defaultRemoteConsent, nil
	}
	return mode, err
}
