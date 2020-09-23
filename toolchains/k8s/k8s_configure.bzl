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

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

_K8S_CONFIGURE_BUILD = """
# DO NOT EDIT: This BUILD file is auto-generated.

package(default_visibility = ["//visibility:public"])

load("@com_adobe_rules_gitops//toolchains/k8s:k8s_toolchain.bzl", "k8s_toolchain")

k8s_toolchain(
    name = "osx_toolchain",
    {kubectl_attr} = "{kubectl}",
)
k8s_toolchain(
    name = "linix_toolchain",
    {kubectl_attr} = "{kubectl}",
)

"""

def _k8s_configure_impl(repository_ctx):
    substitutions = {}
    if repository_ctx.attr.kubectl_path != None:
        substitutions["kubectl_attr"] = "kubectl_path";
        substitutions["kubectl"] = repository_ctx.attr.kubectl_path;
    else:
        substitutions["kubectl_attr"] = "kubectl_target";
        substitutions["kubectl"] = repository_ctx.attr.kubectl_target;

    repository_ctx.file(
        "BUILD",
        content = _K8S_CONFIGURE_BUILD.format(**substitutions),
        executable = False,
    )

_k8s_configure = repository_rule(
    implementation = _k8s_configure_impl,
    attrs = {
        "kubectl_path": attr.label(
            allow_single_file = True,
            mandatory = False,
            doc = "Optional. Path to a prebuilt custom kubectl binary file or label",
        ),
        "kubectl_target": attr.label(
            allow_single_file = True,
            mandatory = False,
            doc = "Optional. Path to a prebuilt custom kubectl binary file or label",
        ),
    },
)

def _ensure_all_provided(func_name, attrs, kwargs):
    """
    Ensures all the required arguments in the given function were specified.

    For function func_name, ensure either all attributes in 'attrs' were
    specified in kwargs or none were specified.
    """
    any_specified = False
    for key in kwargs.keys():
        if key in attrs:
            any_specified = True
            break
    if not any_specified:
        return
    provided = []
    missing = []
    for attr in attrs:
        if attr in kwargs:
            provided.append(attr)
        else:
            missing.append(attr)
    if len(missing) != 0:
        fail("Attribute(s) {} are required for function {} because attribute(s) {} were specified.".format(
            ", ".join(missing),
            func_name,
            ", ".join(provided),
        ))

def k8s_configure(name,
    kubectl_path = None,
    kubectl_target = None
    ):
    """
    Creates an external repository with a configured kubectl_toolchain target.

    Args:
        name: Name of the build target.
        kubectl_path:   Optional. Use the kubectl binary at the given path.
        kubectl_target: Optional. Use the kubectl binary at the given label.

    """
    if "build_srcs" in kwargs and "kubectl_path" in kwargs:
        fail("Attributes 'build_srcs' and 'kubectl_path' can't be specified at" +
             " the same time")
    _k8s_configure(
        name = name,
        kubectl_path = kwargs["kubectl_path"]
    )
