#!/bin/sh
set -eu

sleep 2

i=0
while true; do
  level=info
  if [ $((i % 5)) -eq 0 ]; then
    level=warn
  fi

  payload=$(printf '{"message":"test-app %s log %s","container_name":"test-app","severity":"%s","task_id":"demo","source":"stdout"}' "$level" "$i" "$level")

  curl -sS -X POST http://logfindr:8080/ingest \
    -H 'Content-Type: application/json' \
    -d "$payload" >/dev/null

  i=$((i + 1))
  sleep 1
done
