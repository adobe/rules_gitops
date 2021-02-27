# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.
"""
Defines a repository rule for configuring the Kubernetes tools.
"""

_K8S_TOOLCHAIN_CONFIGURE_BUILD = """
# DO NOT EDIT: This BUILD file is auto-generated.

package(default_visibility = ["//visibility:public"])

load("@com_adobe_rules_gitops//toolchains/k8s:k8s_toolchain.bzl", "k8s_toolchain")

k8s_toolchain(
    name = "toolchain",
    {kubectl_attr} = "{kubectl}",
    {kustomize_attr} = "{kustomize}",
)

"""

def _k8s_toolchain_configure_impl(repository_ctx):
    substitutions = {}

    # configure ether kubectl_path or kubectl_target
    if repository_ctx.attr.kubectl_path != None:
        substitutions["kubectl_attr"] = "kubectl_path"
        substitutions["kubectl"] = repository_ctx.attr.kubectl_path
    else:
        substitutions["kubectl_attr"] = "kubectl_target"
        substitutions["kubectl"] = repository_ctx.attr.kubectl_target

    # configure ether kustoize_path or kustomize_target
    if repository_ctx.attr.kustomize_path != None:
        substitutions["kustomize_attr"] = "kustomize_path"
        substitutions["kustomize"] = repository_ctx.attr.kustomize_path
    else:
        substitutions["kustomize_attr"] = "kustomize_target"
        substitutions["kustomize"] = repository_ctx.attr.kustomize_target
    repository_ctx.file(
        "BUILD",
        content = _K8S_TOOLCHAIN_CONFIGURE_BUILD.format(**substitutions),
        executable = False,
    )

_k8s_toolchain_configure = repository_rule(
    implementation = _k8s_toolchain_configure_impl,
    attrs = {
        "kubectl_path": attr.label(
            allow_single_file = True,
            mandatory = False,
            doc = "Optional. A path to a pre-installed kubectl binary.",
        ),
        "kubectl_target": attr.label(
            allow_single_file = True,
            mandatory = False,
            doc = "Optional. A downloaded or pre-built kubectl binary label",
        ),
        "kustomize_path": attr.label(
            allow_single_file = True,
            mandatory = False,
            doc = "Optional. A path to a pre-installed kustomize binary.",
        ),
        "kustomize_target": attr.label(
            allow_single_file = True,
            mandatory = False,
            doc = "Optional. A downloaded or pre-built kustomize binary label.",
        ),
    },
)

def k8s_toolchain_configure(
        name,
        kubectl_path = None,
        kubectl_target = None):
    """
    Creates an external repository with a configured kubectl_toolchain target.

    Args:
        name: Name of the build target.
        kubectl_path:   Optional. Use the kubectl binary at the given path.
        kubectl_target: Optional. Use the kubectl binary at the given label.

    """
    if "kubectl_path" in kwargs and "kubectl_target" in kwargs:
        fail("Attributes 'kubectl_path' and 'kubectl_target' can't be specified at" +
             " the same time")
    if "kustomize_path" in kwargs and "kustomize_target" in kwargs:
        fail("Attributes 'kustomize_path' and 'kustomize_target' can't be specified at" +
             " the same time")
    _k8s_toolchain_configure(
        name = name,
        kubectl_path = kwargs["kubectl_path"],
    )
