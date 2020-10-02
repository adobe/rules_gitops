# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

load("@bazel_tools//tools/build_rules:test_rules.bzl", "file_test", "rule_test")
load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_image",
)
load("//skylib/kustomize:kustomize.bzl", "gitops", "kubectl", "kustomize", "push_all")
load("//skylib:push.bzl", "k8s_container_push")
load("//skylib:test_rules.bzl", "file_compare_test")

# to generate new test data if needed:
# bazel run //skylib/kustomize:set_namespace newnamespace-1 <test.yaml >test_expected.yaml
sh_test(
    name = "set_namespace_test",
    size = "small",
    srcs = ["set_namespace_test.sh"],
    data = [
        "test.yaml",
        "test_expected.yaml",
        "//skylib/kustomize:set_namespace",
    ],
)

# Verify that rule is combining files without processing
kustomize(
    name = "raw_test",
    testonly = True,
    manifests = [
        "deployment.yaml",
        "service.yaml",
        "crb.yaml",
    ],
    namespace = "",
)

rule_test(
    name = "raw_generates_yaml",
    generates = ["raw_test.yaml"],
    rule = "raw_test",
)

file_compare_test(
    name = "raw_file",
    expected = "expected_raw_test.yaml",
    file = ":raw_test",
)

kustomize(
    name = "namespace_test",
    testonly = True,
    manifests = [
        "deployment.yaml",
        "service.yaml",
        "crb.yaml",
    ],
    namespace = "bs-dev",
)

file_test(
    name = "has_namespace",
    file = ":namespace_test",
    matches = 2,
    regexp = "namespace: bs-dev",
)

filegroup(
    name = "alltestfiles",
    testonly = True,
    srcs = [
        "crb.yaml",
        "deployment.yaml",
        "service.yaml",
    ],
)

kustomize(
    name = "raw2_test",
    testonly = True,
    manifests = [
        ":alltestfiles",
    ],
    namespace = "",
)

file_compare_test(
    name = "raw2_file",
    expected = "expected_raw_test.yaml",
    file = ":raw2_test",
)

container_image(
    name = "image",
    files = ["container_content.txt"],
    visibility = ["//visibility:public"],
)

k8s_container_push(
    name = "image_push",
    image = ":image",
)

kustomize(
    name = "image_test",
    testonly = True,
    images = [
        ":image_push",
    ],
    manifests = [
        "deployment.yaml",
        "service.yaml",
        "crb.yaml",
    ],
    namespace = "",
)

file_compare_test(
    name = "image_resolved_file",
    expected = "expected_image_resolved_test.yaml",
    file = ":image_test",
)

k8s_container_push(
    name = "image_digest_push",
    image = ":image",
    image_digest_tag = True,
)

kustomize(
    name = "image_digest_test",
    testonly = True,
    images = [
        ":image_digest_push",
    ],
    manifests = [
        "deployment.yaml",
        "service.yaml",
        "crb.yaml",
    ],
    namespace = "",
)

file_compare_test(
    name = "image_digest_resolved_file",
    expected = "expected_image_resolved_test.yaml",
    file = ":image_test",
)

k8s_container_push(
    name = "image_push_legacy",
    image = ":image",
    legacy_image_name = "test-image",
)

kustomize(
    name = "legacy_image_test",
    testonly = True,
    images = [
        ":image_push_legacy",
    ],
    manifests = [
        "deployment_legacy.yaml",
        "service.yaml",
        "crb.yaml",
    ],
    namespace = "",
)

file_compare_test(
    name = "logacy_image_resolved_file",
    expected = "expected_image_resolved_test.yaml",
    file = ":legacy_image_test",
)

kustomize(
    name = "configmap_test",
    testonly = True,
    configmaps_srcs = glob(["configmaps/**/*"]),
    namespace = "",
)

file_compare_test(
    name = "configmap_test_file",
    expected = "expected_configmap_test.yaml",
    file = ":configmap_test",
)

kustomize(
    name = "secret_test",
    testonly = True,
    namespace = "",
    secrets_srcs = glob(["secrets/**/*"]),
)

file_compare_test(
    name = "secret_test_file",
    expected = "expected_secret_test.yaml",
    file = ":secret_test",
)

kustomize(
    name = "integration_test",
    testonly = True,
    images = [
        ":image_push",
    ],
    manifests = [
        "deployment.yaml",
        "service.yaml",
        ":configmap_test",
    ],
    namespace = "{BUILD_USER}",
)

# bazel run //skylib/kustomize/tests:push_all
push_all(
    name = "push_all",
    testonly = True,
    srcs = [
        ":integration_test",
    ],
)

kubectl(
    name = "mynamespace.apply",
    testonly = True,
    srcs = [":integration_test"],
    cluster = "dev3",
    namespace = "test",
)

gitops(
    name = "dev-something.gitops",
    testonly = True,
    srcs = [":namespace_test"],
    cluster = "dev",
    namespace = "test",
)

#-------------------
kustomize(
    name = "name_prefix",
    manifests = ["test.yaml"],
    name_prefix = "prefix-",
    namespace = "",
)

file_compare_test(
    name = "name_prefix_test",
    expected = "expected_name_prefix.yaml",
    file = ":name_prefix",
)

#-------------------
kustomize(
    name = "name_suffix",
    manifests = ["test.yaml"],
    name_suffix = "-suffix",
    namespace = "",
)

file_compare_test(
    name = "name_suffix_test",
    expected = "expected_name_suffix.yaml",
    file = ":name_suffix",
)

#-------------------
# patch plus image
kustomize(
    name = "patch",
    images = [
        ":image_push",
    ],
    manifests = ["deployment.yaml"],
    namespace = "",
    patches = ["overlay/deployment.yaml"],
)

file_compare_test(
    name = "patch_test",
    expected = "expected_patch.yaml",
    file = ":patch",
)

#-------------------
# prefix and suffix for deployments: legacy compatibility
kustomize(
    name = "deployment_prefix_compat",
    manifests = ["deployment_with_labels.yaml"],
    name_prefix = "prefix-",
    namespace = "",
)

file_compare_test(
    name = "deployment_prefix_compat_test",
    expected = "expected_deployment_prefix_compat.yaml",
    file = ":deployment_prefix_compat",
)

kustomize(
    name = "deployment_suffix_compat",
    manifests = ["deployment_with_labels.yaml"],
    name_suffix = "-suffix",
    namespace = "",
)

file_compare_test(
    name = "deployment_suffix_compat_test",
    expected = "expected_deployment_suffix_compat.yaml",
    file = ":deployment_suffix_compat",
)

#-------------------
# prefix and suffix for deployments
kustomize(
    name = "deployment_prefix",
    configurations = ["//skylib/kustomize:nameprefix_deployment_labels_config.yaml"],
    manifests = ["deployment_with_labels.yaml"],
    name_prefix = "prefix-",
    namespace = "",
)

file_compare_test(
    name = "deployment_prefix_test",
    expected = "expected_deployment_prefix.yaml",
    file = ":deployment_prefix",
)

kustomize(
    name = "deployment_suffix",
    configurations = ["//skylib/kustomize:nameprefix_deployment_labels_config.yaml"],
    manifests = ["deployment_with_labels.yaml"],
    name_suffix = "-suffix",
    namespace = "",
)

file_compare_test(
    name = "deployment_suffix_test",
    expected = "expected_deployment_suffix.yaml",
    file = ":deployment_suffix",
)

#-------------------
# common labels and common annotations
kustomize(
    name = "common_labels",
    common_annotations = {"ownerTeam": "apps"},
    common_labels = {"flavor": "canary"},
    manifests = [
        "deployment_with_labels.yaml",
        "service.yaml",
    ],
    namespace = "",
)

file_compare_test(
    name = "common_labels_test",
    expected = "expected_common_labels.yaml",
    file = ":common_labels",
)
