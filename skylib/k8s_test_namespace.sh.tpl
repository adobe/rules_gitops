#!/usr/bin/env bash
# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

set -euo pipefail

[ -o xtrace ] && env

function guess_runfiles() {
    pushd ${BASH_SOURCE[0]}.runfiles > /dev/null 2>&1
    pwd
    popd > /dev/null 2>&1
}

RUNFILES=${TEST_SRCDIR:-$(guess_runfiles)}
TEST_UNDECLARED_OUTPUTS_DIR=${TEST_UNDECLARED_OUTPUTS_DIR:-.}

KUBECTL="%{kubectl}"
KUBECONFIG="%{kubeconfig}"
CLUSTER="%{cluster}"
SERVER="%{server}"
USER="%{user}"
BUILD_USER="%{build_user}"

SET_NAMESPACE="%{set_namespace}"
IT_MANIFEST_FILTER="%{it_manifest_filter}"

NAMESPACE_NAME_FILE=${TEST_UNDECLARED_OUTPUTS_DIR}/namespace
KUBECONFIG_FILE=${TEST_UNDECLARED_OUTPUTS_DIR}/kubeconfig

echo "Cluster: ${CLUSTER}" >&2

# use BUILD_USER by defalt
USER=${USER:-$BUILD_USER}

# create miniified self-contained kubectl configuration with the default context set to use newly created namespace
mkdir -p $(dirname $KUBECONFIG_FILE)

# create context partion of new kubeconfig file from scratch
# use --kubeconfig parameter to prevent any merging
# create
rm -f $KUBECONFIG_FILE-context
CONTEXT=$CLUSTER-$BUILD_USER
kubectl --kubeconfig=$KUBECONFIG_FILE-context --cluster=$CLUSTER --server=$SERVER --user=$USER --namespace=$BUILD_USER config  set-context $CONTEXT >&2
kubectl --kubeconfig=$KUBECONFIG_FILE-context config use-context $CONTEXT >&2

# merge newly generated context with system kubeconfig, flatten and minify the result
KUBECONFIG=$KUBECONFIG_FILE-context:$KUBECONFIG kubectl config view --merge=true --minify=true --flatten=true --raw >$KUBECONFIG_FILE

# set generated kubeconfig for all following kubectl commands
export KUBECONFIG=$KUBECONFIG_FILE

# check if username from provided configuration exists
KUBECONFIG_USER=$(${KUBECTL} config view -o jsonpath='{.users[?(@.name == '"\"${USER}\")].name}")
if [ -z "${KUBECONFIG_USER}" ]; then
    echo "Unable to find user configuration ${USER} for cluster ${CLUSTER}" >&2
    exit 1
fi

echo "User: ${USER}" >&2

set +e
if [ -n "${K8S_MYNAMESPACE:-}" ]
then
    # do not create random namesspace
    NAMESPACE=${BUILD_USER}
    # do not delete namespace after the test is complete
    DELETE_NAMESPACE_FLAG=""
    # do not perform manifest transformations
    # test setup should not try to apply modified manifests
    IT_MANIFEST_FILTER="cat"
else
    # create random namespace
    DELETE_NAMESPACE_FLAG="-delete_namespace"
    COUNT="0"
    while true; do
        NAMESPACE=${BUILD_USER}-$(( (RANDOM) + 32767 ))
        ${KUBECTL} create namespace ${NAMESPACE} && break
        COUNT=$[$COUNT + 1]
        if [ $COUNT -ge 10 ]; then
            echo "Unable to create namespace in $COUNT attempts!" >&2
            exit 1
        fi
    done
    # update context with created test namespace
    kubectl --namespace=$NAMESPACE config set-context $CONTEXT >&2

    # rename test context (Note: this is required for backward compatibiliy)
    kubectl config rename-context $CONTEXT $CLUSTER-$NAMESPACE >&2
fi
echo "Namespace: ${NAMESPACE}" >&2
set -e

# expose generated namespace name as rule output
mkdir -p $(dirname $NAMESPACE_NAME_FILE)
echo $NAMESPACE > $NAMESPACE_NAME_FILE

[ -o xtrace ] && kubectl config view >&2

# set runfiles for STMTS
export PYTHON_RUNFILES=${RUNFILES}

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

%{push_statements}
# create k8s objects
%{statements}

%{it_sidecar} -namespace=${NAMESPACE} -timeout=%{test_timeout} %{portforwards} %{waitforapps} ${DELETE_NAMESPACE_FLAG} "$@"
