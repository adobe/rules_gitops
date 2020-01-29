# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

licenses(["notice"])  # Apache 2.0

load(":stamp.bzl", "more_stable_status", "stamp_value")

exports_files([
    "k8s_gitops.sh.tpl",
    "k8s_cmd.sh.tpl",
    "k8s_test_namespace.sh.tpl",
    "push-tag.sh.tpl",
])

stamp_value(
    name = "build_user_value",
    str = "{BUILD_USER}",
    visibility = ["//visibility:public"],
)

more_stable_status(
    name = "more_stable_status",
    vars = [
        "BUILD_USER",
    ],
    visibility = ["//visibility:public"],
)
