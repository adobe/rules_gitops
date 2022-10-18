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
load("@com_adobe_rules_gitops//skylib/kustomize:kustomize.bzl", "kustomize_setup")

def rules_gitops_repositories():
    """Initializes Declares workspaces the GitOps rules depend on.

    Workspaces that use rules_gitops should call this after rules_gitops_dependencies call.
    """

    bazel_skylib_workspace()
    protobuf_deps()
    go_rules_dependencies()
    go_register_toolchains()
    gazelle_dependencies()
    container_repositories()
    container_go_deps()
    kustomize_setup(name = "kustomize_bin")
