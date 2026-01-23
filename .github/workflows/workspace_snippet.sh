#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

TAG=${GITHUB_REF_NAME}
PREFIX="rules_gitops-${TAG:1}"
SHA=$(git archive --format=tar --prefix=${PREFIX}/ ${TAG} | gzip | shasum -a 256 | awk '{print $1}')

cat << EOF

MODULE.bazel snippet:

\`\`\`starlark

bazel_dep(name = "rules_gitops", version = "${TAG}", dev_dependency = True)

\`\`\`

EOF

awk '/---SNIP---/{f=1;next}/---END_SNIP---/{f=0}f' examples/WORKSPACE
echo "\`\`\`"
