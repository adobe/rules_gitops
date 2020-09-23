# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

"""
This module implements the Kubernetes toolchain rule.
"""

K8sInfo = provider(
    doc = "Information about how to invoke the Kuebrnetes tools.",
    fields = {
        "kubectl_path": "Path to the kubectl executable.",
        "kubectl_target": "A kubectl executable target.",
        "kustomize_path": "Path to the kustomize executable.",
        "kustomize_target": "A kustomize executable target.",
        "kubeconfig_path": "Path to kubeconfig file.",
        "kubeconfig_target": "A kubeconfig file target.",
        "kubeconfig_cluster": "Kubernetes cluster name.",
        "kubeconfig_namespace": "Kubernetes namespace.",
        "kubeconfig_user": "Kubernetes user name.",
    },
)

def _k8s_toolchain_impl(ctx):
    if not ctx.attr.kubectl_path and not ctx.attr.kubectl_target:
        fail("No kubectl tool was found.")
    if not ctx.attr.kustomize_path and not ctx.attr.kustomize_target:
        fail("No kustomize tool was found.")
    toolchain_info = platform_common.ToolchainInfo(
        k8s_info = K8sInfo(
            kubectl_path = ctx.attr.kubectl_path,
            kubectl_target = ctx.attr.kubectl_target,
            kustomize_path = ctx.attr.kustomize_path,
            kustomize_target = ctx.attr.kustomize_target,
            kubeconfig_path = ctx.attr.kubeconfig_path,
            kubeconfig_target = ctx.attr.kubeconfig_target,
            kubeconfig_cluster = ctx.attr.kubeconfig_cluster,
            kubeconfig_namespace = ctx.attr.kubeconfig_namespace,
            kubeconfig_user = ctx.attr.kubeconfig_user,
        ),
    )
    return [toolchain_info]

k8s_toolchain = rule(
    implementation = _k8s_toolchain_impl,
    attrs = {
        "kubectl_path": attr.string(
            doc = "Absolute path to a pre-installed kubectl binary.",
            mandatory = False,
        ),
        "kubectl_target": attr.label(
            doc = "Target to downloaded kubectl binary.",
            mandatory = False,
        ),
        "kustomize_path": attr.string(
            doc = "Absolute path to a pre-installed kustomize binary.",
            mandatory = False,
        ),
        "kustomize_target": attr.label(
            doc = "Target to downloaded kustomize binary.",
            mandatory = False,
        ),
        "kubeconfig_path": attr.string(
            doc = "Absolute path to a pre-installed kubeconfig binary.",
            mandatory = False,
        ),
        "kubeconfig_target": attr.label(
            doc = "Target to downloaded kubeconfig binary.",
            mandatory = False,
        ),
        "kubeconfig_cluster": attr.string(
            doc = "Kubernetes cluster name.",
            mandatory = False,
        ),
        "kubeconfig_namespace": attr.string(
            doc = "Kubernetes namespace.",
            mandatory = False,
        ),
        "kubeconfig_user": attr.string(
            doc = "Kubernetes user name.",
            mandatory = False,
        ),
    },
)
