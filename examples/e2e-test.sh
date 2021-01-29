#!/usr/bin/env bash
# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

# This test assume a kind cluster with a registry up and running

set -o errexit
set -o nounset
set -o xtrace

bindir=$(cd `dirname "$0"` && pwd)
repo_path=$bindir
cd $repo_path

#verify interactive workflow
MYNAMESPACE=$USER

# kubectl config use-context kind-kind

kubectl create namespace $MYNAMESPACE || true
kubectl create namespace hwteam || true

bazel run //helloworld:mynamespace.apply
kubectl -n $MYNAMESPACE wait --timeout=60s --for=condition=Available deployment/helloworld

bazel run //helloworld:mynamespace.delete

#TODO: verity it is deleted
#kubectl -n $MYNAMESPACE wait --timeout=30s --for=delete deployment/helloworld

rm -rf cloud

bazel run //helloworld:canary.gitops
bazel run //helloworld:release.gitops
#the result of .gitops operation goes into /cloud directory and should be submitted back to the repo

#apply everything generated
kubectl apply -f cloud -R
#wait for readiness
kubectl -n hwteam wait --timeout=60s --for=condition=Available deployment/helloworld deployment/helloworld-canary
