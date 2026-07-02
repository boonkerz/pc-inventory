package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// SetUserTOTP speichert Secret und Aktivierungsstatus des zweiten Faktors.
func (s *Store) SetUserTOTP(ctx context.Context, userID, secret string, enabled bool) error {
	return s.affect(s.db.ExecContext(ctx, s.rebind(
		`UPDATE users SET totp_secret=?, totp_enabled=? WHERE id=?`), secret, enabled, userID))
}

// ClearUserTOTP deaktiviert 2FA und entfernt Secret + Wiederherstellungscodes.
func (s *Store) ClearUserTOTP(ctx context.Context, userID string) error {
	if _, err := s.db.ExecContext(ctx, s.rebind(`DELETE FROM recovery_codes WHERE user_id=?`), userID); err != nil {
		return err
	}
	return s.affect(s.db.ExecContext(ctx, s.rebind(
		`UPDATE users SET totp_secret='', totp_enabled=FALSE WHERE id=?`), userID))
}

// ReplaceRecoveryCodes ersetzt die Backup-Code-Hashes eines Benutzers.
func (s *Store) ReplaceRecoveryCodes(ctx context.Context, userID string, hashes []string) error {
	if _, err := s.db.ExecContext(ctx, s.rebind(`DELETE FROM recovery_codes WHERE user_id=?`), userID); err != nil {
		return err
	}
	for _, h := range hashes {
		if _, err := s.db.ExecContext(ctx, s.rebind(
			`INSERT INTO recovery_codes (user_id, code_hash, used) VALUES (?, ?, FALSE)`), userID, h); err != nil {
			return err
		}
	}
	return nil
}

// ConsumeRecoveryCode markiert einen gültigen, ungenutzten Backup-Code als benutzt.
func (s *Store) ConsumeRecoveryCode(ctx context.Context, userID, codeHash string) (bool, error) {
	res, err := s.db.ExecContext(ctx, s.rebind(
		`UPDATE recovery_codes SET used=TRUE WHERE user_id=? AND code_hash=? AND used=FALSE`), userID, codeHash)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// CreateLoginChallenge legt eine kurzlebige 2FA-Challenge an (zwischen Passwort und Code).
func (s *Store) CreateLoginChallenge(ctx context.Context, tokenHash, userID string, expires time.Time) error {
	_, err := s.db.ExecContext(ctx, s.rebind(
		`INSERT INTO login_challenges (token_hash, user_id, expires_at) VALUES (?, ?, ?)`),
		tokenHash, userID, expires.UTC())
	return err
}

// ConsumeLoginChallenge liefert die User-ID einer gültigen Challenge und löscht sie.
func (s *Store) ConsumeLoginChallenge(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	var expires time.Time
	err := s.db.QueryRowContext(ctx, s.rebind(
		`SELECT user_id, expires_at FROM login_challenges WHERE token_hash=?`), tokenHash).Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	_, _ = s.db.ExecContext(ctx, s.rebind(`DELETE FROM login_challenges WHERE token_hash=?`), tokenHash)
	if time.Now().After(expires) {
		return "", ErrNotFound
	}
	return userID, nil
}
