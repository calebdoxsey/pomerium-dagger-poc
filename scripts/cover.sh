#!/usr/bin/env bash
# Runs the Go tests with the race detector and writes the coverage profile,
# without generated code and test helpers, to $1.
set -euo pipefail

out="${1:?usage: cover.sh <coverage-profile>}"

# Dagger sets OpenTelemetry variables to collect traces from commands; tests
# that read the OpenTelemetry configuration from the environment expect none.
while read -r name; do
  unset "$name"
done < <(env | sed -n 's/^\(OTEL_[A-Z_]*\)=.*/\1/p')
unset TRACEPARENT

# -coverprofile (unlike -test.gocoverdir) keeps go's test cache working, so only
# packages affected by a change run again. The trade-off: tests rerun by
# --rerun-fails overwrite the profile, so after a flaky failure the coverage
# only reflects the rerun.
gotestsum --rerun-fails --packages=./... -- -race -coverprofile="$out"

sed -i \
  -e '/\.pb\.go:/d' \
  -e '/\/mock_.*\.go:/d' \
  -e '/\.gen\.go:/d' \
  -e '/\/internal\/testenv\//d' \
  -e '/\/internal\/testutil\//d' \
  -e '/\/internal\/tests\//d' \
  -e '/\/storagetest\//d' \
  -e '/\/integration\//d' \
  -e '/\/examples\//d' \
  -e '/\/ssh\/test\//d' \
  "$out"
