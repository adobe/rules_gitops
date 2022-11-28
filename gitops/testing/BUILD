# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("//gitops:defs.bzl", "external_image", "k8s_deploy", "k8s_test_setup")
load("//skylib:push.bzl", "k8s_container_push")
load("//skylib:templates.bzl", "expand_template")

licenses(["notice"])  # Apache 2.0

k8s_deploy(
    name = "legacy_alias",
    cluster = "testcluster",
    deployment_branch = "test1",
    gitops = 1,
    images = {
        "img": "//skylib/kustomize/tests:image",
    },
    manifests = [
        ":deployment_legacy.yaml",
    ],
    namespace = "ci",
    release_branch_prefix = "gitops_test_release_branch",
    visibility = ["//visibility:public"],
)

expand_template(
    name = "generate_deployment_expected",
    out = "deployment_expected.txt",
    deps_aliases = {
        "digest": "//skylib/kustomize/tests:image.digest",
    },
    substitutions = {},
    template = ":deployment_expected_template.txt",
    deps = ["//skylib/kustomize/tests:image.digest"],
)

expand_template(
    name = "generate_deployment2_expected",
    out = "deployment2_expected.txt",
    deps_aliases = {
        "digest": "//skylib/kustomize/tests:image.digest",
    },
    substitutions = {},
    template = ":deployment2_expected_template.txt",
    deps = ["//skylib/kustomize/tests:image.digest"],
)

diff_test(
    name = "legacy_alias_test",
    file1 = ":legacy_alias",
    file2 = "deployment_expected.txt",
)

k8s_deploy(
    name = "legacy_label",
    cluster = "testcluster",
    deployment_branch = "test1",
    gitops = 1,
    images = {
        "img": "//skylib/kustomize/tests:image",
    },
    manifests = [
        ":deployment.yaml",
    ],
    namespace = "ci",
    release_branch_prefix = "gitops_test_release_branch",
    visibility = ["//visibility:public"],
)

diff_test(
    name = "legacy_label_test",
    file1 = ":legacy_label",
    file2 = "deployment_expected.txt",
)

k8s_deploy(
    name = "label",
    cluster = "testcluster",
    deployment_branch = "test1",
    gitops = 1,
    images = [
        "//skylib/kustomize/tests:image",
    ],
    manifests = [
        ":deployment.yaml",
    ],
    namespace = "ci",
    release_branch_prefix = "gitops_test_release_branch",
    visibility = ["//visibility:public"],
)

diff_test(
    name = "label_test",
    file1 = ":label",
    file2 = "deployment_expected.txt",
)

# using image from external repo
external_image(
    name = "external_image",
    digest = "sha:1234567890",
    image = "gcr.io/repo/someimage:thetag",
)

k8s_deploy(
    name = "external_image_label",
    cluster = "testcluster",
    deployment_branch = "test1",
    gitops = 1,
    images = [
        ":external_image",
    ],
    manifests = [
        ":deployment1.yaml",
    ],
    namespace = "ci",
    release_branch_prefix = "gitops_test_release_branch",
    visibility = ["//visibility:public"],
)

diff_test(
    name = "external_image_label_test",
    file1 = ":external_image_label",
    file2 = "deployment1_expected.txt",
)

k8s_deploy(
    name = "external_image_legacy",
    cluster = "testcluster",
    deployment_branch = "test1",
    gitops = 1,
    images = {
        "img": ":external_image",
    },
    manifests = [
        ":deployment_legacy.yaml",
    ],
    namespace = "ci",
    release_branch_prefix = "gitops_test_release_branch",
    visibility = ["//visibility:public"],
)

diff_test(
    name = "external_image_legacy_test",
    file1 = ":external_image_legacy",
    file2 = "deployment1_expected.txt",
)

k8s_container_push(
    name = "pushed_image",
    image = "//skylib/kustomize/tests:image",
    registry = "gcr.io",
    repository = "repo/imagethere",
    tag = "thetag",
)

# rename(
#     name = "//skylib/kustomize/tests:image",

# )

k8s_deploy(
    name = "legacy_renamed_alias",
    cluster = "testcluster",
    deployment_branch = "test1",
    gitops = 1,
    images = {
        "img": ":pushed_image",
    },
    manifests = [
        ":deployment_legacy.yaml",
    ],
    namespace = "ci",
    release_branch_prefix = "gitops_test_release_branch",
    visibility = ["//visibility:public"],
)

diff_test(
    name = "legacy_renamed_alias_test",
    file1 = ":legacy_renamed_alias",
    file2 = "deployment2_expected.txt",
)

k8s_test_setup(
    name = "legacy_renamed_alias_it.setup",
    # kubeconfig = "@k8s_test//:kubeconfig",
    objects = [
        ":legacy_renamed_alias",
    ],
)
