-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS audit_event
(
    id INTEGER PRIMARY KEY,
    type       TEXT  NOT NULL,
    identifier TEXT  NOT NULL,
    message    TEXT  NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE audit_event;