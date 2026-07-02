package store

import (
	"context"
	"time"
)

// SaveBitLockerKey speichert (Upsert) einen Wiederherstellungsschlüssel je Volume.
func (s *Store) SaveBitLockerKey(ctx context.Context, deviceID, mount, recoveryID, key string, now time.Time) error {
	if key == "" {
		return nil // ohne Schlüssel nichts zu escrowen
	}
	res, err := s.db.ExecContext(ctx, s.rebind(`
		UPDATE bitlocker_keys SET recovery_id=?, recovery_key=?, updated_at=? WHERE device_id=? AND mount_point=?`),
		recoveryID, key, now.UTC(), deviceID, mount)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, s.rebind(`
			INSERT INTO bitlocker_keys (device_id, mount_point, recovery_id, recovery_key, updated_at)
			VALUES (?, ?, ?, ?, ?)`), deviceID, mount, recoveryID, key, now.UTC())
	}
	return err
}

// BitLockerKey holt den escrowten Schlüssel eines Volumes (leer, falls keiner).
func (s *Store) BitLockerKey(ctx context.Context, deviceID, mount string) (string, error) {
	var key string
	err := s.db.QueryRowContext(ctx, s.rebind(
		`SELECT recovery_key FROM bitlocker_keys WHERE device_id=? AND mount_point=?`), deviceID, mount).Scan(&key)
	if err != nil {
		return "", err
	}
	return key, nil
}
