#!/usr/bin/env bash
# Cross-compiles every package in every workspace module for Windows (GOARCH
# comes from the environment), except Unix-only server packages and anything
# importing them.
#
# pomerium/cli ships Windows binaries built from these packages, so everything
# except the Unix-only server packages must cross-compile. Only add entries to
# unix_only for code that is genuinely server-only; otherwise fix the build.
# (integration is a test-only package that never builds standalone.)
set -euo pipefail

unix_only='pomerium/(integration|pkg/(ipc|envoy/files))( |$)'

mapfile -t patterns < <(go list -m -f '{{.Dir}}/...')
go list -f '{{if .GoFiles}}{{.ImportPath}} {{join .Deps " "}}{{end}}' "${patterns[@]}" |
  grep -vE "$unix_only" |
  cut -d' ' -f1 |
  xargs go build
