#!/usr/bin/env bash

set -eu

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
set -uo pipefail; set +e; f=bazel_tools/tools/bash/runfiles/runfiles.bash
# shellcheck disable=SC1090
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null || \
source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null || \
source "$0.runfiles/$f" 2>/dev/null || \
source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
{ echo>&2 "ERROR: cannot find $f"; exit 1; }; f=; set -e
# --- end runfiles.bash initialization v3 ---

PIDS=()
function async() {
    # Launch the command asynchronously and track its process id.
    "./$@" &
    PIDS+=($!)
}

function waitpids() {
    # Wait for all of the subprocesses, failing the script if any of them failed.
    if [ "${#PIDS[@]}" != 0 ]; then
        for pid in ${PIDS[@]}; do
            wait ${pid}
        done
    fi
}

%{statements}

