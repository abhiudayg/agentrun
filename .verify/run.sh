#!/bin/sh
set -u

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1

if [ "${1:-all}" != "all" ]; then
  echo "usage: .verify/run.sh [all]" >&2
  exit 2
fi

RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)-$$"
OUT="$ROOT/.verify/artifacts/$RUN_ID"
SUMMARY="$OUT/summary.txt"
mkdir -p "$OUT"

{
  echo "run: $RUN_ID"
  echo "commit: $(git rev-parse HEAD)"
  echo "tree: $(git status --porcelain | wc -l | tr -d ' ') changed paths"
} > "$SUMMARY"

phase() {
  name=$1
  shift
  "$@" > "$OUT/$name.log" 2>&1
  code=$?
  echo "$name exit=$code" >> "$SUMMARY"
  return "$code"
}

if ! phase doctor sh .verify/doctor.sh; then
  echo "doctor exit=blocked" >> "$SUMMARY"
  cat "$SUMMARY"
  exit 3
fi

failed=0
phase unit go test -race -count=1 -cover ./... || failed=1
phase static sh -c 'test -z "$(gofmt -l .)" && go vet ./...' || failed=1
echo "functional exit=not-applicable reason=no HTTP services in this repository" >> "$SUMMARY"

cat "$SUMMARY"
exit "$failed"
