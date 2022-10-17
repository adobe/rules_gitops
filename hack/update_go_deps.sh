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

bindir=$(cd `dirname "$0"` && pwd)
repo_path=$bindir/..
vendordir=$repo_path/vendor

novendor=(
    "golang.org/x/tools"
    "golang.org/x/sync"
    "github.com/golang/glog"
    "github.com/golang/protobuf"
    "github.com/mwitkow/go-proto-validators"
    "github.com/gogo/protobuf"
    "github.com/google/go-genproto"
    "google.golang.org/genproto/"
    "google.golang.org/grpc"
    "github.com/googleapis/googleapis"
    "github.com/bazelbuild/buildtools"
    "github.com/fsnotify/fsnotify"
    "github.com/pelletier/go-toml"
    "github.com/pmezard/go-difflib"
    "github.com/magiconair"
    "github.com/prometheus/client_golang/prometheus/process_collector_windows.go"
    "golang.org/x/crypto/ssh/terminal/util_windows.go"
)

cd $repo_path
GO111MODULE=on go mod vendor
#dep ensure

for pkg in ${novendor[@]}; do
    echo "Removing $pkg..."
    rm -rf ${vendordir}/${pkg}
done

#    -not -iname "*.proto" \
find vendor -type f \
    -not -iname "*.c" \
    -not -iname "*.go" \
    -not -iname "*.h" \
    -not -iname "*.s" \
    -not -iname "AUTHORS*" \
    -not -iname "CONTRIBUTORS*" \
    -not -iname "COPYING*" \
    -not -iname "LICENSE*" \
    -not -iname "NOTICE*" \
    -delete


bazel run //:gazelle
bazel run //:buildifier
