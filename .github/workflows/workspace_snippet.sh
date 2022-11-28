#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

TAG=${GITHUB_REF_NAME}
PREFIX="rules_gitops-${TAG:1}"
SHA=$(git archive --format=tar --prefix=${PREFIX}/ ${TAG} | gzip | shasum -a 256 | awk '{print $1}')

cat << EOF

WORKSPACE snippet:

\`\`\`starlark

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "com_adobe_rules_gitops",
    sha256 = "${SHA}",
    strip_prefix = "${PREFIX}",
    urls = ["https://github.com/adobe/rules_gitops/archive/refs/tags/${TAG}.tar.gz"],
)
EOF

awk '/---SNIP---/{f=1;next}/---END_SNIP---/{f=0}f' examples/WORKSPACE
echo "\`\`\`"
