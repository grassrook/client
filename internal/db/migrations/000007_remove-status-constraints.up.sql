PRAGMA foreign_keys = OFF;

BEGIN TRANSACTION;

CREATE TABLE tournaments_new (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    size INTEGER NOT NULL,
    starting_date INTEGER NOT NULL,
    track_id TEXT NOT NULL REFERENCES tracks(id),
    qualifying_order TEXT NOT NULL,
    randomize_qualifying_order INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'qualifying',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    criteria_config TEXT NOT NULL,
    qualifying_rounds INTEGER NOT NULL,
    CONSTRAINT size_power_of_2
        CHECK (size > 0 AND (size & (size - 1)) = 0)
);

INSERT INTO tournaments_new (
    id, name, size, starting_date, track_id, qualifying_order,
    randomize_qualifying_order, status, created_at, updated_at,
    criteria_config, qualifying_rounds
)
SELECT
    id, name, size, starting_date, track_id, qualifying_order,
    randomize_qualifying_order, status, created_at, updated_at,
    criteria_config, qualifying_rounds
FROM tournaments;

DROP TABLE tournaments;

ALTER TABLE tournaments_new RENAME TO tournaments;

COMMIT;

PRAGMA foreign_keys = ON;