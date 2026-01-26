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

KUBECTL_BIN=$(rlocation "${KUBECTL_BIN_PATH}")
if [[ ! -f "${KUBECTL_BIN}" ]]; then
  echo >&2 "ERROR: could not find kubectl binary"
  exit 1
fi

MYNAMESPACE="${USER}"

echo "=== DEBUG: Cluster status ==="
echo "--- Nodes ---"
"${KUBECTL_BIN}" get nodes -o wide || true

echo "--- All pods in namespace ${MYNAMESPACE} ---"
"${KUBECTL_BIN}" get pods -n "${MYNAMESPACE}" -o wide || true

echo "--- Deployment status ---"
"${KUBECTL_BIN}" get deployments -n "${MYNAMESPACE}" -o wide || true

echo "--- Describe deployment helloworld ---"
"${KUBECTL_BIN}" describe deployment helloworld -n "${MYNAMESPACE}" || true

echo "--- Pod logs (if any) ---"
for pod in $("${KUBECTL_BIN}" get pods -n "${MYNAMESPACE}" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null); do
  echo "=== Logs for pod: ${pod} ==="
  "${KUBECTL_BIN}" logs "${pod}" -n "${MYNAMESPACE}" --tail=50 || true
  echo "=== Events for pod: ${pod} ==="
  "${KUBECTL_BIN}" describe pod "${pod}" -n "${MYNAMESPACE}" | grep -A 20 "^Events:" || true
done

echo "--- Events in namespace ${MYNAMESPACE} ---"
"${KUBECTL_BIN}" get events -n "${MYNAMESPACE}" --sort-by='.lastTimestamp' || true

echo "=== END DEBUG ==="

"${KIND_BIN}" delete cluster -n "${KIND_CLUSTER_NAME}" || true

echo "Deleting kind-registry"

"${DOCKER_BIN}" stop kind-registry || true
"${DOCKER_BIN}" container rm kind-registry || true