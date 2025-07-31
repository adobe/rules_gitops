# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

def _get_path_to_kustomize_on_nixos(repo_ctx):
    """Ephemeral download of kustomize binary."""
    command = "nix-shell --pure -p bash busybox kustomize --run 'which kustomize'"
    result = repo_ctx.execute(
        [
            "sh",
            "-c",
            command,
        ],
        environment = {
            "NIX_PATH": "nixpkgs=/nix/var/nix/profiles/per-user/root/channels/nixos",
        },
    )
    kustomize_downloaded = result.return_code == 0
    if not kustomize_downloaded:
        fail("Failed to run '%s'" % command)

    path_to_kustomize = result.stdout.strip()
    return path_to_kustomize

def is_running_on_nixos(repo_ctx):
    """Check if Bazel is executed on NixOS.

    Args:
        repo_ctx: context of the repository rule,
            containing helper functions and information about attributes

    Returns:
        boolean: indicating if Bazel is executed on NixOS.
    """
    result = repo_ctx.execute([
        "sh",
        "-c",
        "test -f /etc/os-release && cat /etc/os-release | head -n1",
    ])
    os_release_file_read_success = result.return_code == 0
    if not os_release_file_read_success:
        return False

    os_release_first_line = result.stdout
    host_is_nixos = os_release_first_line.strip() == "NAME=NixOS"
    return host_is_nixos

def copy_kustomize_bin_from_nix_store(repo_ctx, path):
    """Copy kustomize binary from nix_store to given path.

    Downloads kustomize binary via nix package manger,
    then copies the binary to Bazel cache. This operation
    does not impact host configuration in any way.

    Args:
        repo_ctx: context of the repository rule,
            containing helper functions and information about attributes
        path: path under which the kustomize should be copied

    Returns:
        None
    """
    repo_ctx.file("%s/kustomize" % path)
    result = repo_ctx.execute([
        "sh",
        "-c",
        "cp -f %s %s/kustomize" % (_get_path_to_kustomize_on_nixos(repo_ctx), path),
    ])
    if result.return_code != 0:
        fail("Failed to copy kustomize bin from nix_store")
