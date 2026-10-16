#!/usr/bin/env bash
# Copyright 2026 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

set -o nounset
set -o pipefail

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

runfiles_export_envvars

DEPLOYMENT_ROOT=""
PERFORM_PUSH="1"
# parse command line parameters
while [[ $# -gt 0 ]]
do
  key="$1"
  case $key in
    -r|--deployment_root|--deployment-root)
    DEPLOYMENT_ROOT="$2"
    shift # past argument
    shift # past value
    ;;
    --nopush)
    PERFORM_PUSH=""
    shift
    ;;
    *)    # unknown option
    echo Unsupported parameter $1
    exit 1
    ;;
  esac
done

# --- begin push helpers ---
PIDS=()
function async() {
    # Launch the command asynchronously and track its process id.
    "$@" &
    PIDS+=($!)
}

function waitpids() {
  # Wait for all subprocesses, returning the exit code of the first failed one.
  local rc=0 r
  if [ "${#PIDS[@]}" != 0 ]; then
    for pid in "${PIDS[@]}"; do
      r=0
      wait "${pid}" || r=$?
      if [ "$rc" == 0 ]; then rc=$r; fi
    done
  fi
  return $rc
}

function retry_push() {
  # Run an image pusher, retrying on any non-zero exit with exponential backoff and jitter.
  local max=${GITOPS_PUSH_MAX_ATTEMPTS:-5}
  local delay=${GITOPS_PUSH_RETRY_DELAY_SECONDS:-5}
  local max_delay=${GITOPS_PUSH_MAX_RETRY_DELAY_SECONDS:-60}
  local n=1 rc
  while true; do
    # errexit may be enabled by the runfiles init, so capture the exit code without tripping it.
    rc=0
    "$@" || rc=$?
    if [ "$rc" == 0 ]; then return 0; fi
    if [ "$n" -ge "$max" ]; then
      echo "push failed after $n attempt(s) (exit code $rc): $*" >&2
      return $rc
    fi
    local sleep_for=$((delay + (delay > 0 ? RANDOM % (delay / 2 + 1) : 0)))
    echo "push attempt $n/$max failed (exit code $rc), retrying in ${sleep_for}s: $*" >&2
    sleep "$sleep_for"
    delay=$((delay * 2))
    if [ "$delay" -gt "$max_delay" ]; then delay=$max_delay; fi
    n=$((n + 1))
  done
}
# --- end push helpers ---

if [ "$PERFORM_PUSH" == "1" ]; then
  %{push_statements}
fi

cd $BUILD_WORKSPACE_DIRECTORY

if [ "%{deployment_branch}" != "" -a "${DEPLOYMENT_ROOT}" != "" ] ; then
  TARGET_DIR=${DEPLOYMENT_ROOT}
else
  echo "--deployment-root or deployment_branch is not specified, using repo root"
  TARGET_DIR=$BUILD_WORKSPACE_DIRECTORY
fi

# make sure that the script exits immediately if any command below fails
set -o errexit
%{statements}
