package store

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/boonkerz/roster/internal/server/model"
)

// ListProxmoxHosts liefert alle Proxmox-Hosts (ohne Token-Secret).
func (s *Store) ListProxmoxHosts(ctx context.Context) ([]model.ProxmoxHost, error) {
	rows, err := s.db.QueryContext(ctx, s.rebind(`
		SELECT id, name, base_url, token_id, verify_tls FROM proxmox_hosts ORDER BY name, base_url`))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ProxmoxHost
	for rows.Next() {
		var h model.ProxmoxHost
		if err := rows.Scan(&h.ID, &h.Name, &h.BaseURL, &h.TokenID, &h.VerifyTLS); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// GetProxmoxHost liefert einen Host inkl. Token-Secret (für serverseitige API-Aufrufe).
func (s *Store) GetProxmoxHost(ctx context.Context, id string) (*model.ProxmoxHost, error) {
	var h model.ProxmoxHost
	err := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, name, base_url, token_id, token_secret, verify_tls FROM proxmox_hosts WHERE id=?`), id).
		Scan(&h.ID, &h.Name, &h.BaseURL, &h.TokenID, &h.TokenSecret, &h.VerifyTLS)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// CreateProxmoxHost legt einen Host an (h.ID muss gesetzt sein).
func (s *Store) CreateProxmoxHost(ctx context.Context, h *model.ProxmoxHost) error {
	_, err := s.db.ExecContext(ctx, s.rebind(`
		INSERT INTO proxmox_hosts (id, name, base_url, token_id, token_secret, verify_tls) VALUES (?, ?, ?, ?, ?, ?)`),
		h.ID, h.Name, h.BaseURL, h.TokenID, h.TokenSecret, h.VerifyTLS)
	return err
}

// DeleteProxmoxHost entfernt einen Host.
func (s *Store) DeleteProxmoxHost(ctx context.Context, id string) error {
	return s.affect(s.db.ExecContext(ctx, s.rebind(`DELETE FROM proxmox_hosts WHERE id=?`), id))
}

// ProxmoxRemediationForCheck liefert die Proxmox-Remediation eines Checks samt Host
// (inkl. Secret), oder ErrNotFound, wenn keine konfiguriert ist.
func (s *Store) ProxmoxRemediationForCheck(ctx context.Context, checkID string) (*model.ProxmoxRemediation, *model.ProxmoxHost, error) {
	checks, err := s.checkByID(ctx, checkID)
	if err != nil {
		return nil, nil, err
	}
	if checks.RemediationProxmox == nil || checks.RemediationProxmox.HostID == "" {
		return nil, nil, ErrNotFound
	}
	host, err := s.GetProxmoxHost(ctx, checks.RemediationProxmox.HostID)
	if err != nil {
		return nil, nil, err
	}
	return checks.RemediationProxmox, host, nil
}

// checkByID lädt einen einzelnen Check (nur die für Remediation nötigen Felder).
func (s *Store) checkByID(ctx context.Context, checkID string) (*model.PolicyCheck, error) {
	var c model.PolicyCheck
	var remProxmox string
	err := s.db.QueryRowContext(ctx, s.rebind(`
		SELECT id, name, remediation_proxmox FROM policy_checks WHERE id=?`), checkID).
		Scan(&c.ID, &c.Name, &remProxmox)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if remProxmox != "" {
		var pr model.ProxmoxRemediation
		if json.Unmarshal([]byte(remProxmox), &pr) == nil && pr.HostID != "" {
			c.RemediationProxmox = &pr
		}
	}
	return &c, nil
}
