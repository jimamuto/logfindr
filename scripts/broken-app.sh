#!/bin/sh
set -eu

i=0
while true; do
  echo "broken-app info log $i"
  if [ $((i % 5)) -eq 0 ]; then
    echo "broken-app error log $i" >&2
  fi
  i=$((i + 1))
  sleep 1
done
