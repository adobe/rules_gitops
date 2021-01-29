#!/usr/bin/env bash
set -euo pipefail

function test_dedupe {
    of=$(mktemp)
    OUTPUT=$(skylib/tests/final.show)
    EXPECTED_OUTPUT=$(cat skylib/tests/fork_join_expected.yaml)
    if [ "${OUTPUT}" != "${EXPECTED_OUTPUT}" ]; then
        echo Unexpected final.show output:
        echo $OUTPUT
        exit 1
    fi
}

test_dedupe
