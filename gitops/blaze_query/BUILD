load("@rules_proto//proto:defs.bzl", "proto_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

licenses(["notice"])  # Apache 2.0

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    embed = [":blaze_query_go_proto"],
    importpath = "github.com/adobe/rules_gitops/gitops/blaze_query",
    visibility = ["//visibility:public"],
)

proto_library(
    name = "blaze_query_proto",
    srcs = ["build.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "blaze_query_go_proto",
    importpath = "github.com/adobe/rules_gitops/gitops/blaze_query",
    proto = ":blaze_query_proto",
    visibility = ["//visibility:public"],
)
