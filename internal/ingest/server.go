package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/logsport/logfindr/internal/db"
)

// Server handles log ingestion from Fluent Bit.
type Server struct {
	db   *db.DB
	addr string
}

// New creates a new ingest server.
func New(database *db.DB, addr string) *Server {
	return &Server{db: database, addr: addr}
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

// Start begins listening for log ingestion.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", s.handleIngest)
	mux.HandleFunc("/health", s.handleHealth)

	fmt.Printf("logfindr ingest server listening on %s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}
