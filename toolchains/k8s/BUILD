# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

load(":k8s_toolchain.bzl", "k8s_toolchain")

package(default_visibility = ["//visibility:private"])

licenses(["notice"])  # Apache 2.0

# Kubernetes toolchain type
toolchain_type(name = "toolchain_type")

# Default kubectl toolchain that expects the 'kubectl' and 'kustomize' executables
# to be in the PATH
k8s_toolchain(
    name = "k8s_default_toolchain",
    kubectl_path = "kubectl",
    kustomize_path = "kustomize",
    visibility = ["//visibility:public"],
)

toolchain(
    name = "k8s_linux_amd64_toolchain",
    target_compatible_with = [
        "@bazel_tools//platforms:linux",
        "@bazel_tools//platforms:x86_64",
    ],
    toolchain = "@k8s_linux_amd64_config//:toolchain",
    toolchain_type = ":toolchain_type",
)

toolchain(
    name = "k8s_darwin_amd64_toolchain",
    target_compatible_with = [
        "@bazel_tools//platforms:osx",
        "@bazel_tools//platforms:x86_64",
    ],
    toolchain = "@k8s_darwin_amd64_config//:toolchain",
    toolchain_type = ":toolchain_type",
)

# Allow targets to use a toolchains attribute, such as sh_binary and genrule
# This way they can reference the make variables e.g KUBECONFIG.
alias(
    name = "toolchain",
    actual = select({
        "@bazel_tools//src/conditions:darwin": "@k8s_darwin_amd64_config//:toolchain",
        "@bazel_tools//src/conditions:linux_x86_64": "@k8s_linux_amd64_config//:toolchain",
        "//conditions:default": ":k8s_default_toolchain",
    }),
    visibility = ["//visibility:public"],
)

# Allow targets to declare a dependency on the kubectl binary for the current host platform
alias(
    name = "kubectl_bin",
    actual = select({
        "@bazel_tools//src/conditions:darwin": "@k8s_darwin_amd64//:kubectl_bin",
        "@bazel_tools//src/conditions:linux_x86_64": "@k8s_linux_amd64//:kubectl_bin",
    }),
    visibility = ["//visibility:public"],
)

# Allow targets to declare a dependency on the kubectl binary for the current host platform
alias(
    name = "kustomize_bin",
    actual = select({
        "@bazel_tools//src/conditions:darwin": "@k8s_darwin_amd64//:kustomize_bin",
        "@bazel_tools//src/conditions:linux_x86_64": "@k8s_linux_amd64//:kustomize_bin",
        "//conditions:default": "@k8s_linux_amd64//:kustomize_bin",
    }),
    visibility = ["//visibility:public"],
)
