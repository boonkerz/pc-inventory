package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/thomaspeterson/pc-inventory/internal/server/model"
	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

// CommandMeta liefert Typ und Payload eines Befehls (für die Ergebnis-Nachbearbeitung).
func (s *Store) CommandMeta(ctx context.Context, id string) (string, map[string]any, error) {
	var typ, payload string
	err := s.db.QueryRowContext(ctx, s.rebind(`SELECT type, payload FROM commands WHERE id=?`), id).Scan(&typ, &payload)
	if err != nil {
		return "", nil, err
	}
	var pl map[string]any
	_ = json.Unmarshal([]byte(payload), &pl)
	return typ, pl, nil
}

// UpsertNetworkAssets fügt gefundene Hosts einer Site hinzu (Dedup über site_id+ip).
func (s *Store) UpsertNetworkAssets(ctx context.Context, siteID string, hosts []shared.NetworkHost) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck
	now := time.Now().UTC()
	n := 0
	for _, h := range hosts {
		var ports []string
		for _, p := range h.Ports {
			ports = append(ports, strconv.Itoa(p))
		}
		if _, err := tx.ExecContext(ctx, s.rebind(`
			INSERT INTO network_assets (id, site_id, ip, mac, hostname, ports, note, first_seen, last_seen)
			VALUES (?, ?, ?, ?, ?, ?, '', ?, ?)
			ON CONFLICT(site_id, ip) DO UPDATE SET
				mac=excluded.mac, hostname=excluded.hostname, ports=excluded.ports, last_seen=excluded.last_seen`),
			newID(), siteID, h.IP, h.MAC, h.Hostname, strings.Join(ports, ","), now, now); err != nil {
			return n, err
		}
		n++
	}
	return n, tx.Commit()
}

// NetworkAssetsForSite liefert die Assets einer Site inkl. „verwaltet"-Flag (ob IP/MAC
// zu einem enrollten Gerät passt).
func (s *Store) NetworkAssetsForSite(ctx context.Context, siteID string) ([]model.NetworkAsset, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT a.id, a.site_id, a.ip, a.mac, a.hostname, a.ports, a.note, a.first_seen, a.last_seen,
			EXISTS(SELECT 1 FROM interfaces i JOIN devices d ON d.id=i.device_id
				WHERE d.revoked=0 AND ((a.mac<>'' AND i.mac=a.mac) OR (a.ip<>'' AND i.ipv4 LIKE '%'||a.ip||'%')))
		FROM network_assets a WHERE a.site_id=? ORDER BY a.ip`), siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.NetworkAsset
	for rows.Next() {
		var a model.NetworkAsset
		if err := rows.Scan(&a.ID, &a.SiteID, &a.IP, &a.MAC, &a.Hostname, &a.Ports, &a.Note, &a.FirstSeen, &a.LastSeen, &a.Managed); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// SetNetworkAssetNote setzt die Notiz eines Assets.
func (s *Store) SetNetworkAssetNote(ctx context.Context, id, note string) error {
	return s.affect(s.db.ExecContext(ctx, s.rebind(`UPDATE network_assets SET note=? WHERE id=?`), note, id))
}

// DeleteNetworkAsset entfernt ein Asset.
func (s *Store) DeleteNetworkAsset(ctx context.Context, id string) error {
	return s.affect(s.db.ExecContext(ctx, s.rebind(`DELETE FROM network_assets WHERE id=?`), id))
}

// kindFromPorts rät den Gerätetyp aus den (CSV-)Ports (für nicht verwaltete Geräte).
func kindFromPorts(csv string) string {
	has := func(ps ...string) bool {
		set := strings.Split(csv, ",")
		for _, p := range ps {
			for _, x := range set {
				if strings.TrimSpace(x) == p {
					return true
				}
			}
		}
		return false
	}
	switch {
	case has("9100", "631", "515"):
		return "Drucker"
	case has("3389"): // RDP ist Windows-spezifisch
		return "Windows"
	case has("22"): // vor SMB: Samba-Server haben oft auch SSH
		return "Linux/SSH"
	case has("445", "139", "135"): // SMB allein ist kein Windows-Beweis (Samba)
		return "Dateiserver (SMB)"
	case has("5900"):
		return "VNC"
	case has("80", "443", "8080"):
		return "Web/Gerät"
	}
	return ""
}

// AdoptNetworkAsset übernimmt ein Asset als nicht verwaltetes Gerät (ohne Agent) und
// entfernt es aus der Asset-Liste. Liefert die neue Geräte-ID.
func (s *Store) AdoptNetworkAsset(ctx context.Context, assetID string) (string, error) {
	var siteID, ip, mac, hostname, ports string
	err := s.db.QueryRowContext(ctx, s.rebind(
		`SELECT site_id, ip, mac, hostname, ports FROM network_assets WHERE id=?`), assetID).
		Scan(&siteID, &ip, &mac, &hostname, &ports)
	if err != nil {
		return "", err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:errcheck
	devID := newID()
	now := time.Now().UTC()
	hn := hostname
	if hn == "" {
		hn = ip
	}
	if _, err := tx.ExecContext(ctx, s.rebind(`
		INSERT INTO devices (id, hostname, os, os_version, first_seen, enrolled_at, agent_token_hash, revoked, site_id, managed)
		VALUES (?, ?, ?, '', ?, ?, '', 0, ?, 0)`),
		devID, hn, kindFromPorts(ports), now, now, siteID); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, s.rebind(`
		INSERT INTO interfaces (device_id, name, mac, ipv4, ipv6) VALUES (?, '', ?, ?, '')`),
		devID, mac, ip); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, s.rebind(`DELETE FROM network_assets WHERE id=?`), assetID); err != nil {
		return "", err
	}
	return devID, tx.Commit()
}

// AdoptAllForSite übernimmt alle noch nicht verwalteten Assets einer Site als Geräte.
func (s *Store) AdoptAllForSite(ctx context.Context, siteID string) (int, error) {
	assets, err := s.NetworkAssetsForSite(ctx, siteID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, a := range assets {
		if a.Managed {
			continue
		}
		if _, err := s.AdoptNetworkAsset(ctx, a.ID); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// MergeUnmanagedDuplicates entfernt nicht verwaltete Platzhalter-Geräte (managed=0),
// die dieselbe MAC oder IPv4 wie das verwaltete Gerät devID haben. Typischer Fall:
// ein per Netzwerk-Scan übernommenes Gerät, auf dem anschließend ein Agent installiert
// wird – dann würden sonst zwei Einträge desselben Hosts nebeneinander stehen. Eine
// evtl. am Platzhalter hinterlegte Notiz wird übernommen, falls das Agent-Gerät keine
// hat. Liefert die Anzahl entfernter Platzhalter.
func (s *Store) MergeUnmanagedDuplicates(ctx context.Context, devID string) (int, error) {
	// MAC- und IPv4-Adressen des verwalteten Geräts sammeln.
	rows, err := s.db.QueryContext(ctx, s.rebind(
		`SELECT mac, ipv4 FROM interfaces WHERE device_id=?`), devID)
	if err != nil {
		return 0, err
	}
	var macs, ips []string
	for rows.Next() {
		var mac, ip sql.NullString
		if err := rows.Scan(&mac, &ip); err != nil {
			rows.Close()
			return 0, err
		}
		if m := strings.ToLower(strings.TrimSpace(mac.String)); m != "" && m != "00:00:00:00:00:00" {
			macs = append(macs, m)
		}
		if v := strings.TrimSpace(ip.String); v != "" {
			ips = append(ips, v)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(macs) == 0 && len(ips) == 0 {
		return 0, nil
	}

	// Kandidaten suchen: managed=0 mit passender MAC oder IPv4 (nicht das Gerät selbst).
	var conds []string
	var args []any
	if len(macs) > 0 {
		in, a := placeholders(macs)
		conds = append(conds, "LOWER(i.mac) IN ("+in+")")
		args = append(args, a...)
	}
	if len(ips) > 0 {
		in, a := placeholders(ips)
		conds = append(conds, "i.ipv4 IN ("+in+")")
		args = append(args, a...)
	}
	q := `SELECT DISTINCT i.device_id FROM interfaces i JOIN devices d ON d.id = i.device_id
		WHERE d.managed = 0 AND i.device_id <> ? AND (` + strings.Join(conds, " OR ") + `)`
	drows, err := s.db.QueryContext(ctx, s.rebind(q), append([]any{devID}, args...)...)
	if err != nil {
		return 0, err
	}
	var dupIDs []string
	for drows.Next() {
		var id string
		if err := drows.Scan(&id); err != nil {
			drows.Close()
			return 0, err
		}
		dupIDs = append(dupIDs, id)
	}
	drows.Close()
	if err := drows.Err(); err != nil {
		return 0, err
	}
	if len(dupIDs) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck
	var myNotes sql.NullString
	_ = tx.QueryRowContext(ctx, s.rebind(`SELECT notes FROM devices WHERE id=?`), devID).Scan(&myNotes)
	for _, dup := range dupIDs {
		if strings.TrimSpace(myNotes.String) == "" {
			var n sql.NullString
			_ = tx.QueryRowContext(ctx, s.rebind(`SELECT notes FROM devices WHERE id=?`), dup).Scan(&n)
			if strings.TrimSpace(n.String) != "" {
				if _, err := tx.ExecContext(ctx, s.rebind(`UPDATE devices SET notes=? WHERE id=?`), n.String, devID); err != nil {
					return 0, err
				}
				myNotes = n
			}
		}
		if _, err := tx.ExecContext(ctx, s.rebind(`DELETE FROM interfaces WHERE device_id=?`), dup); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, s.rebind(`DELETE FROM devices WHERE id=?`), dup); err != nil {
			return 0, err
		}
	}
	return len(dupIDs), tx.Commit()
}
