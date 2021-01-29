#!/usr/bin/env bash
# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

set +x

if [ "$1" == "" ]; then
    echo usage:
    echo $0 'namespace <in.yaml >out.yaml'
    exit 1
fi
set -euo pipefail
dir=$(mktemp -d)
cat >${dir}/in.yaml
cat >${dir}/kustomization.yaml <<EOF
namespace: $1
resources:
- in.yaml
EOF
exec external/kustomize_bin/kustomize build ${dir}
