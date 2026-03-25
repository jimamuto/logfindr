package db

import "time"

// LogEntry represents a single log line stored in the database.
type LogEntry struct {
	ID            int64     `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	TaskID        string    `json:"task_id"`
	Severity      string    `json:"severity"`
	Message       []byte    `json:"-"`        // Zstd compressed
	RawSize       int64     `json:"raw_size"` // original byte count
	Labels        string    `json:"labels"`   // JSON key-value pairs
	Source        string    `json:"source"`   // stdout / stderr
}

const schema = `
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    container_id TEXT,
    container_name TEXT,
    task_id TEXT DEFAULT '',
    severity TEXT DEFAULT 'info',
    message BLOB NOT NULL,
    raw_size INTEGER,
    labels TEXT DEFAULT '{}',
    source TEXT DEFAULT 'stdout'
);

CREATE INDEX IF NOT EXISTS idx_task ON logs(task_id);
CREATE INDEX IF NOT EXISTS idx_container ON logs(container_name);
CREATE INDEX IF NOT EXISTS idx_timestamp ON logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_severity ON logs(severity);

CREATE TABLE IF NOT EXISTS active_task (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    task_id TEXT NOT NULL
);
`
