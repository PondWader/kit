-- Singleton table that stores the last pull (used to decide when to automatically upate repositories)
CREATE TABLE IF NOT EXISTS pull_info (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_pulled_at TEXT,
    last_pull_list_mtime TEXT
) STRICT;

-- Stores packages that are available in the downloaded repositories for fast lookups
CREATE TABLE IF NOT EXISTS packages (
    name TEXT NOT NULL,
    repo TEXT NOT NULL,
    path TEXT NOT NULL,
    PRIMARY KEY (name, repo)
) STRICT;
CREATE INDEX IF NOT EXISTS idx_packages_name ON packages (name);

-- Stores installations of package versions
CREATE TABLE IF NOT EXISTS installations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    repo TEXT NOT NULL,
    version TEXT NOT NULL,
    is_active INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;
CREATE INDEX IF NOT EXISTS idx_installations_name ON installations (name);

-- Stores actions such as symlinks that can be reversed on uninstall/disable
CREATE TABLE IF NOT EXISTS install_mount_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    install_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    data TEXT NOT NULL,
    FOREIGN KEY (id) REFERENCES installations (id)
) STRICT;