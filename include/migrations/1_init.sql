CREATE TABLE IF NOT EXISTS pull_info (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_pulled_at TEXT,
    last_pull_list_mtime TEXT
) STRICT;

CREATE TABLE IF NOT EXISTS installations (
    id INTEGER PRIMARY KEY AUTOINCREMENT
) STRICT;

CREATE TABLE IF NOT EXISTS packages (
    name TEXT NOT NULL,
    repo TEXT NOT NULL,
    path TEXT NOT NULL,
    PRIMARY KEY (name, repo)
) STRICT;
CREATE INDEX IF NOT EXISTS idx_packages_name ON packages (name);