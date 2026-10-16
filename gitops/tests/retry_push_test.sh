#!/usr/bin/env bash
# Copyright 2026 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

# Tests the push helpers (retry_push, async, waitpids) extracted from k8s_gitops.sh.tpl.

set -u

TEMPLATE="${TEST_SRCDIR:-.}/${TEST_WORKSPACE:-.}/gitops/private/k8s_gitops.sh.tpl"
[ -f "$TEMPLATE" ] || TEMPLATE="$(dirname "$0")/../private/k8s_gitops.sh.tpl"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

sed -n '/^# --- begin push helpers ---$/,/^# --- end push helpers ---$/p' "$TEMPLATE" > "$TMP/helpers.sh"
[ -s "$TMP/helpers.sh" ] || { echo "FAIL: push helpers not found in template"; exit 1; }
# shellcheck disable=SC1090
source "$TMP/helpers.sh"

export GITOPS_PUSH_RETRY_DELAY_SECONDS=0

FAILURES=0
function check() {
  if [ "$2" != "$3" ]; then
    echo "FAIL: $1: expected '$2', got '$3'"
    FAILURES=$((FAILURES + 1))
  else
    echo "PASS: $1"
  fi
}

# Fake pusher: fails with exit code 7 for the first $2 invocations, then succeeds.
# Usage: fake_pusher <counter_file> <failures_before_success>
cat > "$TMP/fake_pusher.sh" <<'EOF'
#!/usr/bin/env bash
n=$(cat "$1" 2>/dev/null || echo 0)
n=$((n + 1))
echo "$n" > "$1"
if [ "$n" -le "$2" ]; then exit 7; fi
exit 0
EOF
chmod +x "$TMP/fake_pusher.sh"

function attempts() { cat "$1" 2>/dev/null || echo 0; }

# succeeds first time
rm -f "$TMP/c1"
retry_push "$TMP/fake_pusher.sh" "$TMP/c1" 0 2>/dev/null
check "success without retry: rc" 0 $?
check "success without retry: attempts" 1 "$(attempts "$TMP/c1")"

# transient failures are retried until success
rm -f "$TMP/c2"
retry_push "$TMP/fake_pusher.sh" "$TMP/c2" 3 2>/dev/null
check "recovers after transient failures: rc" 0 $?
check "recovers after transient failures: attempts" 4 "$(attempts "$TMP/c2")"

# persistent failure gives up after max attempts and preserves the exit code
rm -f "$TMP/c3"
retry_push "$TMP/fake_pusher.sh" "$TMP/c3" 100 2>/dev/null
check "gives up: rc preserved" 7 $?
check "gives up: attempts" 5 "$(attempts "$TMP/c3")"

# GITOPS_PUSH_MAX_ATTEMPTS=1 disables retries
rm -f "$TMP/c4"
GITOPS_PUSH_MAX_ATTEMPTS=1 retry_push "$TMP/fake_pusher.sh" "$TMP/c4" 100 2>/dev/null
check "retries disabled: rc" 7 $?
check "retries disabled: attempts" 1 "$(attempts "$TMP/c4")"

# custom max attempts
rm -f "$TMP/c5"
GITOPS_PUSH_MAX_ATTEMPTS=2 retry_push "$TMP/fake_pusher.sh" "$TMP/c5" 100 2>/dev/null
check "custom max: attempts" 2 "$(attempts "$TMP/c5")"

# parallel pushers: one flaky recovers, all succeed
rm -f "$TMP/p1" "$TMP/p2"
PIDS=()
async retry_push "$TMP/fake_pusher.sh" "$TMP/p1" 0
async retry_push "$TMP/fake_pusher.sh" "$TMP/p2" 2
waitpids 2>/dev/null
check "parallel flaky recovers: rc" 0 $?
check "parallel flaky recovers: attempts" 3 "$(attempts "$TMP/p2")"

# parallel pushers: one persistently failing makes waitpids fail, others still complete
rm -f "$TMP/q1" "$TMP/q2" "$TMP/q3"
PIDS=()
async retry_push "$TMP/fake_pusher.sh" "$TMP/q1" 100
async retry_push "$TMP/fake_pusher.sh" "$TMP/q2" 0
async retry_push "$TMP/fake_pusher.sh" "$TMP/q3" 1
waitpids 2>/dev/null
check "parallel persistent failure: rc" 7 $?
check "parallel persistent failure: others complete" "1 2" "$(attempts "$TMP/q2") $(attempts "$TMP/q3")"

if [ "$FAILURES" != 0 ]; then
  echo "$FAILURES check(s) failed"
  exit 1
fi
echo "All checks passed"
