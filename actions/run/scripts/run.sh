#!/usr/bin/env bash
# Runs `dagger VERB ARGS...` with the CLI in DIR, and puts it on the PATH for
# later steps.
#
# Inputs (environment): DIR, VERB, ARGS. ARGS is split on whitespace (without
# globbing) into separate arguments, so arguments can't contain spaces.
set -euo pipefail

echo "$DIR" >>"$GITHUB_PATH"
"$DIR/dagger" version

# read returns non-zero at the end of its input, which -d '' always reaches.
read -r -d '' -a dagger_args <<<"$ARGS" || true
"$DIR/dagger" "$VERB" "${dagger_args[@]}"
