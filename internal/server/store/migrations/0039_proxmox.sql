-- +goose Up
-- Proxmox-Hosts: Zugang zur Proxmox-VE-API (API-Token). Zugangsdaten bleiben
-- serverseitig; der Server ruft die API im Auftrag der UI/Remediation auf.
CREATE TABLE proxmox_hosts (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL DEFAULT '',
    base_url     TEXT NOT NULL,                    -- z.B. https://pve.local:8006
    token_id     TEXT NOT NULL,                    -- user@realm!tokenid
    token_secret TEXT NOT NULL DEFAULT '',         -- Token-Secret (UUID)
    verify_tls   BOOLEAN NOT NULL DEFAULT FALSE,   -- Proxmox nutzt meist selbstsigniertes Zert.
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Proxmox-Remediation je Check: bei „failing" rebootet der Server den hinterlegten
-- Gast (LXC/QEMU). Leer = keine. JSON {"host_id","node","type","vmid"}.
ALTER TABLE policy_checks ADD COLUMN remediation_proxmox TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE policy_checks DROP COLUMN remediation_proxmox;
DROP TABLE proxmox_hosts;
