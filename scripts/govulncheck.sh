#!/usr/bin/env bash
# Runs govulncheck on every module in the Go workspace.
set -euo pipefail

for dir in $(go list -m -f '{{.Dir}}'); do
  echo "==> govulncheck $dir"
  (cd "$dir" && govulncheck ./...)
done
