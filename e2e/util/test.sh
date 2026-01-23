#!/usr/bin/env bash

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
set -uo pipefail; set +e; f=bazel_tools/tools/bash/runfiles/runfiles.bash
# shellcheck disable=SC1090
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null || \
source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null || \
source "$0.runfiles/$f" 2>/dev/null || \
source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
{ echo>&2 "ERROR: cannot find $f"; exit 1; }; f=; set -e
# --- end runfiles.bash initialization v3 ---

# Track overall exit code
exit_code=0

# Resolve teardown script path (needed for trap)
teardown_script=""
if [[ -n "${TEARDOWN:-}" ]]; then
  teardown_script=$(rlocation "${TEARDOWN}")
  if [[ -z "${teardown_script}" || ! -f "${teardown_script}" ]]; then
    echo >&2 "ERROR: could not find TEARDOWN script at rlocationpath: ${TEARDOWN}"
    exit 1
  fi
fi

# Teardown runs on exit, regardless of success or failure
cleanup() {
  if [[ -n "${teardown_script}" ]]; then
    echo "=== Running teardown: ${TEARDOWN} ==="
    "${teardown_script}" || true
    echo "=== Teardown completed ==="
  fi
}
trap cleanup EXIT

# Run setup if specified
if [[ -n "${SETUP:-}" ]]; then
  setup_script=$(rlocation "${SETUP}")
  if [[ -z "${setup_script}" || ! -f "${setup_script}" ]]; then
    echo >&2 "ERROR: could not find SETUP script at rlocationpath: ${SETUP}"
    exit 1
  fi

  echo "=== Running setup: ${SETUP} ==="
  if ! "${setup_script}"; then
    echo >&2 "=== Setup failed ==="
    exit 1
  fi
  echo "=== Setup completed ==="
fi

# Run all scripts passed as arguments
for rlocation_path in "$@"; do
  script=$(rlocation "${rlocation_path}")
  if [[ -z "${script}" || ! -f "${script}" ]]; then
    echo >&2 "ERROR: could not find script at rlocationpath: ${rlocation_path}"
    exit_code=1
    break
  fi

  echo "=== Running: ${rlocation_path} ==="
  if ! "${script}"; then
    echo >&2 "=== Failed: ${rlocation_path} ==="
    exit_code=1
    break
  fi
  echo "=== Completed: ${rlocation_path} ==="
done

if [[ ${exit_code} -eq 0 ]]; then
  echo "=== All scripts completed successfully ==="
fi

exit ${exit_code}
