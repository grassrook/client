CREATE TABLE qualifying_scores (
    id TEXT PRIMARY KEY NOT NULL,
    round INT NOT NULL,
    tournament_id TEXT NOT NULL
        REFERENCES tournaments(id) ON DELETE CASCADE,
    participant_id TEXT NOT NULL
        REFERENCES participants(id) ON DELETE CASCADE,
    score REAL NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    UNIQUE (tournament_id, participant_id, round)
);

INSERT INTO qualifying_scores (
    id,
    round,
    tournament_id,
    participant_id,
    score,
    status,
    created_at,
    updated_at
)
SELECT
    s.id,
    s.round,
    p.tournament_id,
    s.participant_id,
    0,
    'pending',
    s.created_at,
    s.updated_at
FROM scores s
JOIN participants p ON p.id = s.participant_id
WHERE s.kind = 'qualifying';

DROP TABLE IF EXISTS scores;