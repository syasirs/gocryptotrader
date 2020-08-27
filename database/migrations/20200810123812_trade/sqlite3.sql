-- +goose Up
CREATE TABLE IF NOT EXISTS trade
(
    id  TEXT PRIMARY KEY NOT NULL UNIQUE ON CONFLICT REPLACE,
    tid TEXT,
    exchange_id TEXT NOT NULL,
    currency TEXT NOT NULL,
    asset TEXT NOT NULL,
    price real NOT NULL,
    amount real NOT NULL,
    side TEXT NOT NULL,
    timestamp real NOT NULL
);
-- +goose Down
DROP TABLE trade;
