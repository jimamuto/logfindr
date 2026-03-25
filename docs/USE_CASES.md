# Use Cases

Detailed examples of how logfindr fits into real developer workflows.

## Container Logging

Point any Docker container's logs to logfindr using the fluentd log driver. No code changes needed — just add the logging config to your compose file.

```yaml
services:
  my-app:
    image: my-app:latest
    logging:
      driver: fluentd
      options:
        fluentd-address: localhost:24224
        tag: my-app
```

All logs from `my-app` will flow into logfindr, compressed and indexed automatically.

## AI Agent Debugging

Coding agents (Claude Code, Cursor, Copilot Workspace) can pull error context directly instead of relying on you to copy-paste terminal output.

```bash
# Agent runs this to understand what went wrong
./logfindr query --task fix-auth-bug --severity error

# Agent compares a broken run against a stable one
./logfindr compare --task-a fix-auth-bug --task-b stable-v2
```

This turns logfindr into a persistent memory bank for autonomous coding workflows. The agent doesn't need a browser, a dashboard, or your help — it queries the logs itself.

## Task-Based Debugging

Tag your current work session. All incoming logs get labeled with that task ID automatically. When you switch tasks, just re-tag.

```bash
# Start working on a bug
./logfindr tag --task JIRA-1234

# ... debug, restart containers, reproduce the issue ...

# Switch to a different task
./logfindr tag --task JIRA-5678

# Later, pull logs from either task — they're all still there
./logfindr query --task JIRA-1234
./logfindr query --task JIRA-5678
```

No more scrolling through interleaved logs trying to figure out which output belongs to which debugging session.

## Regression Comparison

Compare a broken deploy against a stable one to spot exactly what changed in the error patterns.

```bash
./logfindr compare --task-a deploy-v2.1 --task-b deploy-v2.0
```

Output:

```
Task Comparison: deploy-v2.1 vs deploy-v2.0
============================================
                     deploy-v2.1 deploy-v2.0
Total logs                   847         612
Errors                        23           2
Warnings                      41          10

--- Errors in deploy-v2.1 ---
  [14:32:01] payment-service: connection pool exhausted
  [14:32:03] payment-service: timeout waiting for db connection
  [14:32:05] api-gateway: upstream service unavailable
  ...
```

Immediately tells you: v2.1 introduced a database connection pool issue that cascaded into gateway errors.

## Post-Mortem Auditing

Logs persist in SQLite on a Docker volume. Containers crash, restart, get redeployed — the logs survive all of it.

```bash
# What errors happened in the payment service last week?
./logfindr query --container payment-service --since 168h --severity error

# How many total errors across all containers in the last 24h?
./logfindr query --since 24h --severity error
```

No more "we lost the logs when the container restarted." They're on disk, compressed, and queryable indefinitely.

## Direct API Ingestion

Any application can POST logs directly to logfindr — not just Docker containers. Useful for scripts, cron jobs, CI runners, or anything that can make an HTTP call.

```bash
curl -X POST http://localhost:8080/ingest \
  -d '{
    "message": "order #4521 failed: insufficient inventory",
    "severity": "error",
    "task_id": "checkout-fix",
    "container_name": "order-processor",
    "labels": "{\"order_id\": \"4521\", \"env\": \"staging\"}"
  }'
```

This means logfindr works as a general-purpose log sink, not just a Docker tool.

## CI/CD Pipeline Logging

Tag logs per pipeline run, then query failures after the fact. Useful when CI output scrolls past and you need to find a specific error.

```bash
# At the start of a CI run
./logfindr tag --task ci-run-4521

# Run your test suite, deploy scripts, etc.
# All container logs during this window are tagged with ci-run-4521

# After the run, pull just the errors
./logfindr query --task ci-run-4521 --severity error
```

Works especially well when your CI pipeline spins up multiple containers (database, API, workers) and you need to correlate errors across them.

## Multi-Service Debugging

When debugging across multiple microservices, filter by container to isolate each service's logs, or query by task to see the full picture.

```bash
# See what the API gateway logged during your debugging session
./logfindr query --task fix-timeout --container api-gateway

# See what the database proxy logged at the same time
./logfindr query --task fix-timeout --container db-proxy

# See everything together
./logfindr query --task fix-timeout
```

## Storage Monitoring

Keep an eye on how much space your logs are consuming and how well compression is working.

```bash
./logfindr stats
```

```
Logfindr Statistics
====================
Total logs:        12847
Total tasks:       23
DB file size:      4.2 MB
Raw log data:      38.7 MB
Stored (Zstd):     4.1 MB
Compression ratio: 9.4x
```

38.7 MB of raw logs stored in 4.2 MB on disk. Months of history without filling your drive.
