#!/usr/bin/env bash

# Debug
# set -x
# RUNFILES_LIB_DEBUG=1

# --- begin runfiles.bash initialization v2 ---
# Copy-pasted from the Bazel Bash runfiles library v2.
set -uo pipefail
f=bazel_tools/tools/bash/runfiles/runfiles.bash
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null ||
  source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null ||
  source "$0.runfiles/$f" 2>/dev/null ||
  source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null ||
  source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null ||
  {
    echo >&2 "ERROR: cannot find $f"
    exit 1
  }
f=
set -e
# --- end runfiles.bash initialization v2 ---

CLUSTER="$1"
NAMESPACE="$2"

$(rlocation examples/helloworld/mynamespace.show) > mynamespace.show
echo "DEBUG: mynamespace.show:"
cat mynamespace.show

grep -F "kind: Deployment" mynamespace.show
grep -F "kind: Service" mynamespace.show
grep -F "name: helloworld" mynamespace.show
grep -E "image: localhost:5000/.*/helloworld/image@sha256" mynamespace.show
grep -E "app_label_image_short_digest" mynamespace.show | grep -v -F 'image.short-digest'

$(rlocation examples/helloworld/canary.show) > canary.show
echo "DEBUG: canary.show:"
cat canary.show

grep -F "kind: Deployment" canary.show
grep -F "kind: Service" canary.show
grep -F "namespace: $NAMESPACE" canary.show
grep -F "name: helloworld-canary" canary.show
grep -E "image: localhost:5000/k8s/helloworld/image@sha256" canary.show

$(rlocation examples/helloworld/release.show) > release.show
echo "DEBUG: release.show:"
cat release.show

grep -F "kind: Deployment" release.show
grep -F "kind: Service" release.show
grep -F "namespace: $NAMESPACE" canary.show
grep -F "name: helloworld" mynamespace.show
grep -E "image: localhost:5000/k8s/helloworld/image@sha256" release.show
