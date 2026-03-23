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

DOCKER_BIN=$(which docker)
if [[ ! -f "${DOCKER_BIN}" ]]; then
  echo >&2 "ERROR: could not find docker binary"
  exit 1
fi

echo "Using docker: ${DOCKER_BIN}"

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='15000'
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker container rm "${reg_name}" 2>/dev/null || true
  docker run \
    -d --restart=always -e "REGISTRY_HTTP_ADDR=0.0.0.0:${reg_port}" -p "${reg_port}:${reg_port}" --name "${reg_name}" \
    registry:3
fi

# create a cluster with the local registry enabled in containerd
cat <<EOF | "${KIND_BIN}" create cluster --name "${KIND_CLUSTER_NAME}" --image "kindest/node:v1.30.13" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
EOF

# connect the registry to the cluster network
# (the network may already be connected)
docker network connect "${KIND_CLUSTER_NAME}" "${reg_name}" || true

"${KUBECTL_BIN}" config use-context kind-kind

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | "${KUBECTL_BIN}" apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF



MYNAMESPACE=$USER

echo "Creating namespaces ${MYNAMESPACE} and hwteam"

"${KUBECTL_BIN}" create namespace $MYNAMESPACE || true
"${KUBECTL_BIN}" create namespace hwteam || true