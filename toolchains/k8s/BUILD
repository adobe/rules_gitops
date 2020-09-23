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
    name = "default_k8s",
    kubectl_path = "kubectl",
    kustomize_path = "kustomize",
    visibility = ["//visibility:public"],
)

toolchain(
    name = "k8s_linux_toolchain",
    target_compatible_with = [
        "@bazel_tools//platforms:linux",
        "@bazel_tools//platforms:x86_64",
    ],
    toolchain = "@k8s_config//:toolchain",
    toolchain_type = ":toolchain_type",
)

toolchain(
    name = "k8s_osx_toolchain",
    target_compatible_with = [
        "@bazel_tools//platforms:osx",
        "@bazel_tools//platforms:x86_64",
    ],
    toolchain = "@k8s_config//:toolchain",
    toolchain_type = ":toolchain_type",
)

