#!/bin/sh
set -eu

i=0
while true; do
  echo "stable-app info log $i"
  if [ $((i % 7)) -eq 0 ]; then
    echo "stable-app warn log $i" >&2
  fi
  i=$((i + 1))
  sleep 1
done
