#!/usr/bin/env bash
# Generate a CA, a server certificate for the Dagger engine, and a client
# certificate for CI. Usage: ./gen-certs.sh <engine-hostname> [output-dir]
set -euo pipefail

host="${1:?usage: $0 <engine-hostname> [output-dir]}"
out="${2:-./certs}"
mkdir -p "$out"
cd "$out"
umask 077

# CA. Keep ca.key offline; it is only needed to issue or rotate certificates.
openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout ca.key -out ca.crt -days 3650 -subj "/CN=dagger-engine-ca"

# Server certificate, valid for the hostname the CI runners connect to.
openssl req -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout tls.key -out tls.csr -subj "/CN=${host}"
openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out tls.crt -days 825 \
  -extfile <(printf 'subjectAltName=DNS:%s\nextendedKeyUsage=serverAuth\n' "$host")

# Client certificate for GitHub Actions.
openssl req -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout client.key -out client.csr -subj "/CN=github-actions"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 825 \
  -extfile <(printf 'extendedKeyUsage=clientAuth\n')

rm -f ./*.csr ca.srl
echo "Wrote $(pwd): ca.crt ca.key tls.crt tls.key client.crt client.key"
