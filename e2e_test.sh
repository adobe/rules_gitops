#!/usr/bin/env bash
# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

set -o errexit
set -o nounset
set -o xtrace

bindir=$(cd `dirname "$0"` && pwd)
repo_path=$bindir
cd $repo_path

#check installs
bazel version
docker version
which kubectl
go version

go get sigs.k8s.io/kind@v0.11.1

cluster_running="$(docker inspect -f '{{.State.Running}}' kind-control-plane 2>/dev/null || true)"
if [ "${cluster_running}" != 'true' ]; then
  ./create_kind_cluster.sh
fi

delete() {
    echo "Cleanup..."
}

# Setup a trap to delete the namespace on error
set +o xtrace
trap "echo FAILED ; delete" EXIT
set -o xtrace

./examples/e2e-test.sh

delete

# Replace the exit trap with a pass message
trap "echo PASS" EXIT
