#!/usr/bin/env bash
set -euo pipefail

function test_namespace_replaced {
  OUTPUT=`cat skylib/kustomize/tests/test.yaml | skylib/kustomize/set_namespace newnamespace-1`
  EXPECTED_OUTPUT=$(cat skylib/kustomize/tests/test_expected.yaml)
  if [ "${OUTPUT}" != "${EXPECTED_OUTPUT}" ]; then
    echo Unexpected set_namespace output:
    echo $OUTPUT
    exit 1
  fi
}

test_namespace_replaced
