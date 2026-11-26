#!/usr/bin/env bash
# Points later steps at the shared Dagger engine by setting DAGGER_ENGINE.
#
# Inputs (environment): DAGGER_ADDR, DAGGER_CA_CERT, DAGGER_CLIENT_CERT,
# DAGGER_CLIENT_KEY. Without credentials (for example on pull requests from
# forks) it does nothing, and Dagger starts a temporary local engine.
set -euo pipefail

if [[ -z "${DAGGER_CLIENT_KEY:-}" ]]; then
  echo "No engine credentials; using a local engine."
  exit 0
fi

d="$RUNNER_TEMP/dagger-tls"
mkdir -p "$d"
printf '%s\n' "$DAGGER_CA_CERT" >"$d/ca.crt"
printf '%s\n' "$DAGGER_CLIENT_CERT" >"$d/client.crt"
(umask 077 && printf '%s\n' "$DAGGER_CLIENT_KEY" >"$d/client.key")
echo "DAGGER_ENGINE=tls://$DAGGER_ADDR?ca_path=$d/ca.crt&cert_path=$d/client.crt&key_path=$d/client.key" >>"$GITHUB_ENV"
