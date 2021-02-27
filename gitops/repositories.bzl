# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

"""
GtiOps rules repositories initialization
"""

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")
load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load("@io_bazel_rules_docker//repositories:repositories.bzl", container_repositories = "repositories")
load("@io_bazel_rules_docker//repositories:go_repositories.bzl", container_go_deps = "go_deps")
load("@com_adobe_rules_gitops//toolchains/k8s:k8s_toolchain_configure.bzl", "k8s_toolchain_configure")

# Supported OS and Architecture combinations
_OS_ARCH_NAMES = [
    ("darwin", "amd64"),
    ("linux", "amd64"),
]

def rules_gitops_toolchains(
        toolchain_repositories_prefix = "k8s",
        register_toolchains = True,
        **kwargs):
    """Setups Kubernets support toolchains for different platforms

    Args:
        toolchain_repositories_prefix: A toolchains reposittories name prefix.
        register_toolchains: register default toolchains.
        **kwargs: arguments passed to k8s_repositories_rule.
    """

    # This needs to be setup so toolchains can access kubernetes tools
    for os_arch_name in OS_ARCH_NAMES:
        os_name = "_".join(os_arch_name)  # {os}_{arch}
        k8s_repository_name = "k8s_%s" % os_name
        _maybe(
            k8s_repositories_rule,
            name = k8s_repository_name,
            **kwargs
        )
        if register_toolchains:
            # Register toolchain defined in toolchains/k8s/BUILD that reference @k8s_{os}_{arch}_config//:toolchain repositores
            native.register_toolchains("@com_adobe_rules_gitops//toolchains/k8s:k8s_%s_toolchain" % os_name)
            k8s_toolchain_configure(
                name = "%s_config" % k8s_repository_name,
                kubectl_target = "@%s//:kubectl_bin" % k8s_repository_name,
                kuestomize_target = "@%s//:kustomize_bin" % k8s_repository_name,
            )

    # replaces
    # kustomize_setup(name = "kustomize_bin")

def rules_gitops_repositories(**kwargs):
    """Initializes Declares workspaces the GitOps rules depend on.

    Workspaces that use rules_gitops should call this after rules_gitops_dependencies call.

    Args:
        **kwargs: arguments passed to rules_gitops_toolchains macro
    """

    bazel_skylib_workspace()
    protobuf_deps()
    go_rules_dependencies()
    go_register_toolchains()
    gazelle_dependencies()
    container_repositories()
    container_go_deps()
    rules_gitops_toolchains(**kwargs)

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
