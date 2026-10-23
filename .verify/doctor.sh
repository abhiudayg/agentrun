#!/bin/sh
fail=0
for tool in go gofmt git; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "missing tool: $tool"
    fail=1
  fi
done
exit "$fail"
