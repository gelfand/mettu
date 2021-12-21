#!/bin/sh
while true; do
  go run ./cmd/website
  $@ &
  PID=$!
  inotifywait -r -e modify .
  kill $PID
done
