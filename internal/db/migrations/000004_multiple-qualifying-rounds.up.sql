ALTER TABLE tournaments ADD COLUMN qualifying_rounds INT DEFAULT 1; 

CREATE TABLE qualifying_scores {
    id TEXT PRIMARY KEY NOT NULL,
    round INT NOT NULL,
    tournament_id TEXT NOT NULL
        REFERENCES tournaments(id) ON DELETE CASCADE,
    participant_id TEXT NOT NULL 
        REFERENCES participants(id) ON DELETE CASCADE,
    score REAL NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    UNIQUE (tournament_id, participant_id, round)
};