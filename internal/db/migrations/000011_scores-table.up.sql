CREATE TABLE scores(
  id TEXT PRIMARY KEY NOT NULL,
  tournament_id TEXT
      REFERENCES tournaments(id) ON DELETE CASCADE,
  participant_id TEXT
      REFERENCES participants(id) ON DELETE CASCADE,
  match_id TEXT DEFAULT NULL,
  round INTEGER DEFAULT 1,
  kind TEXT NOT NULL,
  results TEXT NULL CHECK (json_valid(results)),
  score REAL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

INSERT INTO scores (
    id,
    participant_id,
    match_id,
    round,
    score,
    kind,
    results,
    created_at,
    updated_at
)
SELECT
    id,
    participant_id,
    NULL,
    round,
    score,
    'qualifying',
    '{}',
    created_at,
    updated_at
FROM qualifying_scores;

DROP TABLE qualifying_scores;