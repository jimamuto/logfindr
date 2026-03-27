package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/logsport/logfindr/internal/compress"
	_ "modernc.org/sqlite"
)

// DB wraps the SQLite connection.
type DB struct {
	conn *sql.DB
}

// Open creates or opens the SQLite database in WAL mode.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	conn.SetMaxOpenConns(1)
	if _, err := conn.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

// Insert stores a log entry with Zstd-compressed message.
func (d *DB) Insert(entry *LogEntry) error {
	raw := entry.Message
	compressed := compress.Compress(raw)
	_, err := d.conn.Exec(
		`INSERT INTO logs (timestamp, container_id, container_name, task_id, severity, message, raw_size, labels, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.Timestamp.UTC(), entry.ContainerID, entry.ContainerName,
		entry.TaskID, entry.Severity, compressed, len(raw), entry.Labels, entry.Source,
	)
	return err
}

// QueryFilter holds optional filters for querying logs.
type QueryFilter struct {
	TaskID    string
	Container string
	Severity  string
	Since     time.Duration
	Limit     int
}

// Query retrieves log entries matching the filter.
func (d *DB) Query(f QueryFilter) ([]LogEntry, error) {
	q := "SELECT id, timestamp, container_id, container_name, task_id, severity, message, raw_size, labels, source FROM logs WHERE 1=1"
	var args []interface{}

	if f.TaskID != "" {
		q += " AND task_id = ?"
		args = append(args, f.TaskID)
	}
	if f.Container != "" {
		q += " AND container_name = ?"
		args = append(args, f.Container)
	}
	if f.Severity != "" {
		q += " AND severity = ?"
		args = append(args, f.Severity)
	}
	if f.Since > 0 {
		q += " AND timestamp >= ?"
		args = append(args, time.Now().Add(-f.Since).UTC())
	}

	q += " ORDER BY timestamp DESC"

	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	q += " LIMIT ?"
	args = append(args, limit)

	rows, err := d.conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		var ts string
		if err := rows.Scan(&e.ID, &ts, &e.ContainerID, &e.ContainerName, &e.TaskID, &e.Severity, &e.Message, &e.RawSize, &e.Labels, &e.Source); err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse("2006-01-02 15:04:05+00:00", ts)
		if e.Timestamp.IsZero() {
			e.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", ts)
		}
		// Decompress message
		decoded, err := compress.Decompress(e.Message)
		if err != nil {
			return nil, fmt.Errorf("decompress log %d: %w", e.ID, err)
		}
		e.Message = decoded
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ListTasks returns all distinct task IDs with their log counts.
func (d *DB) ListTasks() ([]TaskInfo, error) {
	rows, err := d.conn.Query(
		`SELECT task_id, COUNT(*) as cnt, MIN(timestamp) as first_seen, MAX(timestamp) as last_seen
		 FROM logs WHERE task_id != '' GROUP BY task_id ORDER BY last_seen DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TaskInfo
	for rows.Next() {
		var t TaskInfo
		if err := rows.Scan(&t.TaskID, &t.Count, &t.FirstSeen, &t.LastSeen); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// TaskInfo holds summary info about a task's logs.
type TaskInfo struct {
	TaskID    string `json:"task_id"`
	Count     int64  `json:"count"`
	FirstSeen string `json:"first_seen"`
	LastSeen  string `json:"last_seen"`
}

// Stats returns database statistics.
type Stats struct {
	TotalLogs        int64   `json:"total_logs"`
	TotalTasks       int64   `json:"total_tasks"`
	DBSizeBytes      int64   `json:"db_size_bytes"`
	TotalRawBytes    int64   `json:"total_raw_bytes"`
	TotalStoredBytes int64   `json:"total_stored_bytes"`
	CompressionRatio float64 `json:"compression_ratio"`
}

func (d *DB) Stats(dbPath string) (*Stats, error) {
	var s Stats
	d.conn.QueryRow("SELECT COUNT(*) FROM logs").Scan(&s.TotalLogs)
	d.conn.QueryRow("SELECT COUNT(DISTINCT task_id) FROM logs WHERE task_id != ''").Scan(&s.TotalTasks)
	d.conn.QueryRow("SELECT COALESCE(SUM(raw_size), 0) FROM logs").Scan(&s.TotalRawBytes)
	d.conn.QueryRow("SELECT COALESCE(SUM(LENGTH(message)), 0) FROM logs").Scan(&s.TotalStoredBytes)

	if s.TotalStoredBytes > 0 {
		s.CompressionRatio = float64(s.TotalRawBytes) / float64(s.TotalStoredBytes)
	}

	fi, err := os.Stat(dbPath)
	if err == nil {
		s.DBSizeBytes = fi.Size()
	}
	return &s, nil
}

// ContainerInfo holds summary info about a container's logs.
type ContainerInfo struct {
	ContainerName string `json:"container_name"`
	Count         int64  `json:"count"`
	LastSeen      string `json:"last_seen"`
}

// ListContainers returns all distinct container names with their log counts.
func (d *DB) ListContainers() ([]ContainerInfo, error) {
	rows, err := d.conn.Query(
		`SELECT container_name, COUNT(*) as cnt, MAX(timestamp) as last_seen
		 FROM logs WHERE container_name != '' GROUP BY container_name ORDER BY last_seen DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []ContainerInfo
	for rows.Next() {
		var c ContainerInfo
		if err := rows.Scan(&c.ContainerName, &c.Count, &c.LastSeen); err != nil {
			return nil, err
		}
		containers = append(containers, c)
	}
	return containers, rows.Err()
}

// SetActiveTask sets the current active task ID for tagging incoming logs.
func (d *DB) SetActiveTask(taskID string) error {
	_, err := d.conn.Exec(
		`INSERT INTO active_task (id, task_id) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET task_id = ?`,
		taskID, taskID,
	)
	return err
}

// GetActiveTask returns the current active task ID, or empty string if none.
func (d *DB) GetActiveTask() string {
	var taskID string
	d.conn.QueryRow("SELECT task_id FROM active_task WHERE id = 1").Scan(&taskID)
	return taskID
}
