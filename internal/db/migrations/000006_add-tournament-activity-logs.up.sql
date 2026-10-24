CREATE TABLE tournament_activity_logs(
  id TEXT PRIMARY KEY NOT NULL,
  tournament_id TEXT NOT NULL
      REFERENCES tournaments(id) ON DELETE CASCADE,
  log_type TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);