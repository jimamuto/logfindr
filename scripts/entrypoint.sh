#!/bin/sh
set -e

echo "Starting logfindr ingest server..."
/usr/local/bin/logfindr serve --db /data/logfindr.db --addr :8080 &
LOGFINDR_PID=$!

sleep 1

echo "Starting Fluent Bit..."
/opt/fluent-bit/bin/fluent-bit -c /etc/fluent-bit/fluent-bit.conf &
FLUENTBIT_PID=$!

trap "kill $LOGFINDR_PID $FLUENTBIT_PID 2>/dev/null; exit 0" SIGTERM SIGINT

wait $LOGFINDR_PID $FLUENTBIT_PID
