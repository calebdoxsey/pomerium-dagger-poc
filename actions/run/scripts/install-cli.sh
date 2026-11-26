#!/usr/bin/env bash
# Downloads the Dagger CLI release archive, verifies it against the published
# checksums and installs the binary into DIR.
#
# Inputs (environment): DAGGER_VERSION, PLATFORM, DIR.
set -euo pipefail

archive="dagger_${DAGGER_VERSION}_${PLATFORM}.tar.gz"
base="https://dl.dagger.io/dagger/releases/${DAGGER_VERSION#v}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

curl -fsSL --retry 3 -o "$tmp/$archive" "$base/$archive"
curl -fsSL --retry 3 -o "$tmp/checksums.txt" "$base/checksums.txt"
(cd "$tmp" && grep " $archive\$" checksums.txt | sha256sum -c -)

mkdir -p "$DIR"
tar -xzf "$tmp/$archive" -C "$DIR" dagger
