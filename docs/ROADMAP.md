# Roadmap

Sprint plan and goals for logfindr MVP and beyond.

## Sprint 1: Core Foundation (Done)

Delivered the working foundation.

- [x] Go binary with Cobra CLI
- [x] SQLite storage layer with WAL mode
- [x] Zstd compression for log payloads
- [x] HTTP ingest API (POST /ingest, GET /health)
- [x] Fluent Bit integration via forward input
- [x] CLI commands: serve, query, tasks, compare, tag, stats
- [x] Multi-stage Dockerfile
- [x] docker-compose.yml for one-command deployment
- [x] GitHub Actions CI pipeline pushing to Docker Hub
- [x] MIT license

## Sprint 2: MVP Hardening

Make it production-ready for solo developers and small teams.

- [ ] JSON output mode (`--output json`) for all CLI commands — machine-readable output for agents
- [ ] Log retention policy — auto-delete logs older than N days (`--retain 30d`)
- [ ] Batch ingest endpoint — accept arrays of log entries in a single POST
- [ ] Graceful shutdown — drain in-flight requests before stopping
- [ ] Health check improvements — include DB status, Fluent Bit status, uptime
- [ ] Error handling hardening — retry logic for DB writes, better error messages
- [ ] Container auto-discovery labels — parse Docker container labels into log metadata

## Sprint 3: Agent Integration

Make logfindr the standard way coding agents get log context.

- [ ] MCP (Model Context Protocol) server — expose logfindr as a tool that Claude Code, Cursor, and other agents can call natively
- [ ] Structured log parsing — auto-detect JSON logs and extract fields for richer queries
- [ ] Full-text search — search within log message content, not just metadata filters
- [ ] Log streaming — `logfindr tail --task <id>` for real-time log following
- [ ] Export command — `logfindr export --task <id> --format json` to dump logs for external tools

## Sprint 4: Multi-User and Scale

Support small teams sharing a logfindr instance.

- [ ] Authentication — API key-based auth for the ingest and query endpoints
- [ ] Multi-user task namespacing — isolate tasks per user/team
- [ ] Remote CLI — query a logfindr instance running on another machine
- [ ] Webhook notifications — trigger alerts when error count exceeds a threshold
- [ ] Dashboard — minimal web UI for browsing logs (optional, keep CLI-first)

## Sprint 5: Ecosystem

Grow the project and integrations.

- [ ] Helm chart for Kubernetes deployment
- [ ] GitHub Action for logfindr — ingest CI logs automatically
- [ ] VS Code extension — query logs from the editor sidebar
- [ ] Grafana datasource plugin — visualize logfindr data in Grafana dashboards
- [ ] Plugin system — custom log parsers and output formatters

## MVP Definition

The MVP is complete when Sprint 1 + Sprint 2 are done. At that point logfindr:

1. Runs as a single Docker container pulled from Docker Hub
2. Ingests logs from any container via fluentd log driver
3. Stores logs persistently with compression
4. Lets you query, compare, and audit logs by task
5. Outputs JSON for agent consumption
6. Auto-cleans old logs to prevent unbounded disk growth
7. Handles failures gracefully without data loss
