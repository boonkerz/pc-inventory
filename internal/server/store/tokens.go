package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/thomaspeterson/pc-inventory/internal/server/model"
)

// CreateEnrollmentToken speichert ein Enrollment-Token (bereits gehasht).
func (s *Store) CreateEnrollmentToken(ctx context.Context, t *model.EnrollmentToken, tokenHash string) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	var exp, site any
	if t.ExpiresAt != nil {
		exp = t.ExpiresAt.UTC()
	}
	if t.SiteID != nil {
		site = *t.SiteID
	}
	_, err := s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO enrollment_tokens (id, label, token_hash, expires_at, max_uses, used_count, created_by, created_at, site_id)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)`),
		t.ID, t.Label, tokenHash, exp, t.MaxUses, t.CreatedBy, t.CreatedAt, site)
	return err
}

// EnsureEnrollmentToken legt ein Enrollment-Token (unbegrenzt nutzbar) an, falls
// noch keines mit diesem Hash existiert. Für automatisierte Test-Umgebungen.
func (s *Store) EnsureEnrollmentToken(ctx context.Context, label, tokenHash, createdBy string) error {
	var n int
	if err := s.db.QueryRowContext(ctx, s.rebind(
		`SELECT COUNT(*) FROM enrollment_tokens WHERE token_hash = ?`), tokenHash).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return s.CreateEnrollmentToken(ctx, &model.EnrollmentToken{
		ID:        newID(),
		Label:     label,
		MaxUses:   0, // unbegrenzt
		CreatedBy: createdBy,
	}, tokenHash)
}

// ListEnrollmentTokens liefert alle Tokens (ohne Klartext).
func (s *Store) ListEnrollmentTokens(ctx context.Context) ([]model.EnrollmentToken, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, label, expires_at, max_uses, used_count, created_by, created_at
		FROM enrollment_tokens ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.EnrollmentToken
	for rows.Next() {
		var t model.EnrollmentToken
		var exp sql.NullTime
		if err := rows.Scan(&t.ID, &t.Label, &exp, &t.MaxUses, &t.UsedCount, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		if exp.Valid {
			e := exp.Time
			t.ExpiresAt = &e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// DeleteEnrollmentToken widerruft ein Enrollment-Token.
func (s *Store) DeleteEnrollmentToken(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, s.rebind(`DELETE FROM enrollment_tokens WHERE id = ?`), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ConsumeEnrollmentToken prüft Gültigkeit (Ablauf + max. Nutzungen) und erhöht used_count atomar.
// Liefert die optional am Token hinterlegte Site-ID zurück (für die Auto-Zuordnung).
// Gibt ErrNotFound bzw. ErrTokenExhausted zurück, wenn das Token nicht (mehr) nutzbar ist.
func (s *Store) ConsumeEnrollmentToken(ctx context.Context, tokenHash string) (*string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	var id string
	var exp sql.NullTime
	var maxUses, used int
	var siteID sql.NullString
	err = tx.QueryRowContext(ctx, s.rebind(`
		SELECT id, expires_at, max_uses, used_count, site_id FROM enrollment_tokens WHERE token_hash = ?`),
		tokenHash).Scan(&id, &exp, &maxUses, &used, &siteID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if exp.Valid && time.Now().After(exp.Time) {
		return nil, ErrTokenExhausted
	}
	if maxUses > 0 && used >= maxUses {
		return nil, ErrTokenExhausted
	}
	if _, err := tx.ExecContext(ctx, s.rebind(`
		UPDATE enrollment_tokens SET used_count = used_count + 1 WHERE id = ?`), id); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	var site *string
	if siteID.Valid {
		v := siteID.String
		site = &v
	}
	return site, nil
}

// ErrTokenExhausted signalisiert ein abgelaufenes oder aufgebrauchtes Token.
var ErrTokenExhausted = errors.New("enrollment-token abgelaufen oder aufgebraucht")
