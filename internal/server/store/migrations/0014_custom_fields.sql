-- +goose Up
CREATE TABLE custom_fields (
    id            TEXT PRIMARY KEY,
    model         TEXT NOT NULL,             -- client | site | device
    name          TEXT NOT NULL,             -- case-sensitiver Bezeichner
    type          TEXT NOT NULL,             -- text|number|checkbox|select|multiselect|datetime|list
    options       TEXT NOT NULL DEFAULT '[]', -- JSON-Array (für select/multiselect)
    default_value TEXT NOT NULL DEFAULT '',
    required      BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(model, name)
);

CREATE TABLE custom_field_values (
    field_id  TEXT NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    entity_id TEXT NOT NULL,                 -- id des client/site/device
    value     TEXT NOT NULL DEFAULT '',      -- list/multiselect: JSON-Array, sonst String
    PRIMARY KEY (field_id, entity_id)
);
CREATE INDEX idx_cfv_entity ON custom_field_values(entity_id);

-- Tasks, die ihre JSON-Ausgabe in Felder übernehmen (Collector).
ALTER TABLE policy_tasks ADD COLUMN collect_fields BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE policy_tasks DROP COLUMN collect_fields;
DROP TABLE custom_field_values;
DROP TABLE custom_fields;
