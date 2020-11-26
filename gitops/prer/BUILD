# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

licenses(["notice"])  # Apache 2.0

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["create_gitops_prs.go"],
    importpath = "github.com/adobe/rules_gitops/gitops/prer",
    visibility = ["//visibility:private"],
    deps = [
        "//gitops/analysis:go_default_library",
        "//gitops/bazel:go_default_library",
        "//gitops/commitmsg:go_default_library",
        "//gitops/exec:go_default_library",
        "//gitops/git:go_default_library",
        "//gitops/git/bitbucket:go_default_library",
        "//gitops/git/github:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
    ],
)

go_binary(
    name = "create_gitops_prs",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
