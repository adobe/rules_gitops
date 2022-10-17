# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

# gazelle:build_tags darwin,linux
# gazelle:exclude examples

licenses(["notice"])  # Apache 2.0

load("@bazel_gazelle//:def.bzl", "gazelle")
load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

exports_files(["WORKSPACE"])

# gazelle:prefix github.com/adobe/rules_gitops
gazelle(
    name = "gazelle",
    build_tags = [
        "integration",
        "debug",
    ],
    command = "fix",
    extra_args = [
        "-build_file_name",
        "BUILD,BUILD.bazel",
    ],
)

buildifier(
    name = "buildifier",
    lint_mode = "warn",
    lint_warnings = [
        "-module-docstring",
        "-function-docstring",
        "-function-docstring-header",
        "-function-docstring-args",
        "-function-docstring-return",
    ],
)

buildifier(
    name = "buildifier-fix",
    lint_mode = "fix",
)

buildifier(
    name = "buildifier-check",
    lint_mode = "warn",
    lint_warnings = [
        "-module-docstring",
        "-function-docstring",
        "-function-docstring-header",
        "-function-docstring-args",
        "-function-docstring-return",
    ],
    mode = "check",
)
