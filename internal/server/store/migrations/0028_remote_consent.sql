-- +goose Up
-- Zustimmungs-Modus für die Fernsteuerung, zielgruppenabhängig (Vererbung
-- device -> site -> client, Default unbeaufsichtigt).
CREATE TABLE remote_consent (
    target_type TEXT NOT NULL, -- device | site | client
    target_id   TEXT NOT NULL,
    mode        TEXT NOT NULL, -- unattended | prompt
    PRIMARY KEY (target_type, target_id)
);

-- +goose Down
DROP TABLE remote_consent;
