package ingest

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/logsport/logfindr/internal/db"
)

//go:embed dashboard.html
var dashboardHTML []byte

// Server handles log ingestion from Fluent Bit.
type Server struct {
	db     *db.DB
	dbPath string
	addr   string
}

// New creates a new ingest server.
func New(database *db.DB, addr string, dbPath string) *Server {
	return &Server{db: database, addr: addr, dbPath: dbPath}
}

type ingestPayload struct {
	Timestamp     string `json:"timestamp"`
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	TaskID        string `json:"task_id"`
	Severity      string `json:"severity"`
	Message       string `json:"message"`
	Labels        string `json:"labels"`
	Source        string `json:"source"`
}

func (s *Server) insertPayload(p ingestPayload) error {
	ts := time.Now().UTC()
	if p.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, p.Timestamp); err == nil {
			ts = parsed
		}
	}

	taskID := p.TaskID
	if taskID == "" {
		taskID = s.db.GetActiveTask()
	}

	severity := p.Severity
	if severity == "" {
		severity = "info"
	}

	labels := p.Labels
	if labels == "" {
		labels = "{}"
	}

	source := p.Source
	if source == "" {
		source = "stdout"
	}

	entry := &db.LogEntry{
		Timestamp:     ts,
		ContainerID:   p.ContainerID,
		ContainerName: p.ContainerName,
		TaskID:        taskID,
		Severity:      severity,
		Message:       []byte(p.Message),
		Labels:        labels,
		Source:        source,
	}

	return s.db.Insert(entry)
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
		return
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		http.Error(w, "bad json: empty body", http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(trimmed, "[") {
		var payloads []ingestPayload
		if err := json.Unmarshal(body, &payloads); err != nil {
			http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
			return
		}

		for _, p := range payloads {
			if err := s.insertPayload(p); err != nil {
				http.Error(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		var p ingestPayload
		if err := json.Unmarshal(body, &p); err != nil {
			http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.insertPayload(p); err != nil {
			http.Error(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy"}`)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dashboardHTML)
}

type logResponse struct {
	ID            int64  `json:"id"`
	Timestamp     string `json:"timestamp"`
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	TaskID        string `json:"task_id"`
	Severity      string `json:"severity"`
	Message       string `json:"message"`
	RawSize       int64  `json:"raw_size"`
	Labels        string `json:"labels"`
	Source        string `json:"source"`
}

func (s *Server) handleAPILogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 100
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	var since time.Duration
	if v := q.Get("since"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			since = d
		}
	}

	entries, err := s.db.Query(db.QueryFilter{
		TaskID:    q.Get("task"),
		Container: q.Get("container"),
		Severity:  q.Get("severity"),
		Since:     since,
		Limit:     limit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]logResponse, len(entries))
	for i, e := range entries {
		resp[i] = logResponse{
			ID:            e.ID,
			Timestamp:     e.Timestamp.Format(time.RFC3339),
			ContainerID:   e.ContainerID,
			ContainerName: e.ContainerName,
			TaskID:        e.TaskID,
			Severity:      e.Severity,
			Message:       string(e.Message),
			RawSize:       e.RawSize,
			Labels:        e.Labels,
			Source:        e.Source,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleAPITasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.ListTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) handleAPIContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := s.db.ListContainers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.Stats(s.dbPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Start begins listening for log ingestion.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", s.handleIngest)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/logs", s.handleAPILogs)
	mux.HandleFunc("/api/tasks", s.handleAPITasks)
	mux.HandleFunc("/api/stats", s.handleAPIStats)
	mux.HandleFunc("/api/containers", s.handleAPIContainers)
	mux.HandleFunc("/", s.handleDashboard)

	fmt.Printf("logfindr ingest server listening on %s\n", s.addr)
	fmt.Printf("dashboard available at http://localhost%s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}
