package store

import (
	"context"
	"time"
)

// MetricPoint ist ein (aggregierter) Punkt der Auslastungs-Historie.
type MetricPoint struct {
	TS   int64   `json:"ts"` // Unix-Millisekunden (Bucket-Beginn)
	CPU  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk"`
}

// InsertMetricsSample speichert eine Auslastungs-Momentaufnahme.
func (s *Store) InsertMetricsSample(ctx context.Context, deviceID string, tsMillis int64, cpu, mem, disk float64) error {
	_, err := s.db.ExecContext(ctx, s.rebind(
		`INSERT INTO metrics_samples (device_id, ts, cpu, mem, disk) VALUES (?, ?, ?, ?, ?)`),
		deviceID, tsMillis, cpu, mem, disk)
	return err
}

// MetricsHistory liefert die Historie ab sinceMillis, auf Zeit-Buckets (bucketMillis)
// gemittelt.
func (s *Store) MetricsHistory(ctx context.Context, deviceID string, sinceMillis, bucketMillis int64) ([]MetricPoint, error) {
	if bucketMillis <= 0 {
		bucketMillis = 5 * 60 * 1000
	}
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT (ts/?)*? AS bucket, AVG(cpu), AVG(mem), AVG(disk)
		FROM metrics_samples WHERE device_id=? AND ts>=?
		GROUP BY ts/? ORDER BY bucket`),
		bucketMillis, bucketMillis, deviceID, sinceMillis, bucketMillis)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MetricPoint
	for rows.Next() {
		var p MetricPoint
		if err := rows.Scan(&p.TS, &p.CPU, &p.Mem, &p.Disk); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// PruneMetrics löscht Samples älter als retention.
func (s *Store) PruneMetrics(ctx context.Context, retention time.Duration) error {
	if retention <= 0 {
		return nil
	}
	cutoff := time.Now().Add(-retention).UnixMilli()
	_, err := s.db.ExecContext(ctx, s.rebind(`DELETE FROM metrics_samples WHERE ts < ?`), cutoff)
	return err
}
