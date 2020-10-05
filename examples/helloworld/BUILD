# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

licenses(["notice"])  # Apache 2.0

load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")
load("@com_adobe_rules_gitops//gitops:defs.bzl", "k8s_deploy")

go_library(
    name = "go_default_library",
    srcs = ["helloworld.go"],
    importpath = "github.com/adobe/rules_gitops/examples/helloworld",
    visibility = ["//visibility:private"],
)

go_test(
    name = "go_default_test",
    srcs = ["helloworld_test.go"],
    embed = [":go_default_library"],
)

go_binary(
    name = "helloworld",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_image(
    name = "image",
    embed = [":go_default_library"],
    goarch = "amd64",
    goos = "linux",
    visibility = ["//visibility:public"],
)

CLUSTER = "kind-kind"

USER = "kind-kind"

REGISTRY = "localhost:5000"

k8s_deploy(
    name = "mynamespace",
    cluster = CLUSTER,
    image_registry = REGISTRY,  # override the default registry host for developent
    images = {
        "helloworld-image": ":image",
    },
    manifests = [
        "deployment.yaml",
        "service.yaml",
    ],
    namespace = "{BUILD_USER}",
    user = USER,
)

NAMESPACE = "hwteam"

k8s_deploy(
    name = "canary",
    cluster = CLUSTER,
    deployment_branch = "helloworld-canary",
    image_digest_tag = True,  # test optional image tagging
    image_registry = REGISTRY,  # override the default registry for production
    image_repository_prefix = "k8s",
    images = {
        "helloworld-image": ":image",
    },
    manifests = [
        "deployment.yaml",
        "service.yaml",
    ],
    name_suffix = "-canary",
    namespace = NAMESPACE,
    prefix_suffix_app_labels = True,
    user = USER,
)

k8s_deploy(
    name = "release",
    cluster = CLUSTER,
    deployment_branch = "helloworld-prod",
    image_digest_tag = True,  # test optional image tagging
    image_registry = REGISTRY,  # override the default registry host for production
    image_repository_prefix = "k8s",
    images = {
        "helloworld-image": ":image",
    },
    manifests = [
        "deployment.yaml",
        "service.yaml",
    ],
    namespace = NAMESPACE,
    user = USER,
)

sh_test(
    name = "k8s_deploy_test",
    srcs = ["k8s_deploy_test.sh"],
    args = [
        CLUSTER,
        NAMESPACE,
    ],
    data = [
        ":canary.show",
        ":mynamespace.show",
        ":release.show",
    ],
    deps = [
        "@bazel_tools//tools/bash/runfiles",
    ],
)
