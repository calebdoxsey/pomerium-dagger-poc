#!/usr/bin/env bash
# Picks the Dagger CLI build for this runner and where it is installed.
#
# Inputs (environment): DAGGER_VERSION, RUNNER_OS, RUNNER_ARCH, TOOL_CACHE.
# Outputs: version, platform, dir.
set -euo pipefail

case "$RUNNER_OS/$RUNNER_ARCH" in
  Linux/X64) platform=linux_amd64 ;;
  Linux/ARM64) platform=linux_arm64 ;;
  *)
    echo "::error::unsupported runner: $RUNNER_OS/$RUNNER_ARCH"
    exit 1
    ;;
esac

{
  echo "version=$DAGGER_VERSION"
  echo "platform=$platform"
  echo "dir=$TOOL_CACHE/dagger/$DAGGER_VERSION/$platform"
} >>"$GITHUB_OUTPUT"
