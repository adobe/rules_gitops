#!/usr/bin/env bash
#
set -euo pipefail

bindir=$(cd `dirname "$0"` && pwd)
repo_path=$bindir/..
cd $repo_path

bazel build //gitops/blaze_query/... //gitops/analysis/...
GEN_REPB_PATH="gitops/blaze_query/build.pb.go"
cp -f "$(find $(bazel info bazel-bin) -path "*/$GEN_REPB_PATH")" "$GEN_REPB_PATH"

GEN_REPB_PATH="gitops/analysis/analysis.pb.go"
cp -f "$(find $(bazel info bazel-bin) -path "*/$GEN_REPB_PATH")" "$GEN_REPB_PATH"

