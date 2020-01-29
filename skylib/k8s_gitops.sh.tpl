#!/bin/bash
# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

set -o nounset
set -o pipefail

is_bazel_run=true
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
    --nobazel)
    is_bazel_run=false
    shift
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

# find_repo_root prints the absolute path to the repository root. This is
# discovered by locating the WORKSPACE file in a parent directory, and
# deferencing symlinks. If no WORKSPACE file is found, this function prints
# an error and exits.
function find_repo_root {
  pushd . &>/dev/null
  while [ $(pwd) != / -a ! -f WORKSPACE ]; do
    cd ..
  done
  if [ ! -f WORKSPACE ]; then
    echo "error: could not find WORKSPACE file in any parent directory" >&2
    exit 1
  fi
  if [ "$is_bazel_run" = true ]; then
    dirname $(readlink WORKSPACE)
  else
    pwd
  fi
  popd &>/dev/null
}

function guess_runfiles() {
    pushd ${BASH_SOURCE[0]}.runfiles > /dev/null 2>&1
    pwd
    popd > /dev/null 2>&1
}

RUNFILES="${PYTHON_RUNFILES:-$(guess_runfiles)}"

PIDS=()
function async() {
    # Launch the command asynchronously and track its process id.
    PYTHON_RUNFILES=${RUNFILES} "$@" &
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

REPO_ROOT=$(find_repo_root)
cd $REPO_ROOT

if [ "%{deployment_branch}" != "" -a "${DEPLOYMENT_ROOT}" != "" ] ; then
  TARGET_DIR=${DEPLOYMENT_ROOT}
else
  echo "--deployment-root or deployment_branch is not specified, using repo root"
  TARGET_DIR=$REPO_ROOT
fi

%{statements}
