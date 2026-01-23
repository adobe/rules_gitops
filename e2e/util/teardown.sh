#!/usr/bin/env bash
set -o errexit

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

DOCKER_BIN=$(which docker)
if [[ ! -f "${DOCKER_BIN}" ]]; then
  echo >&2 "ERROR: could not find docker binary"
  exit 1
fi

# desired cluster name; default is "kind"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"

KIND_BIN=$(rlocation "${KIND_BIN_PATH}")
if [[ ! -f "${KIND_BIN}" ]]; then
  echo >&2 "ERROR: could not find kind binary"
  exit 1
fi

"${KIND_BIN}" delete cluster -n "${KIND_CLUSTER_NAME}" || true

echo "Deleting kind-registry"

"${DOCKER_BIN}" stop kind-registry || true
"${DOCKER_BIN}" container rm kind-registry || true