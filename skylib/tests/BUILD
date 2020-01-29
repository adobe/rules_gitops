# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

load("//skylib:k8s.bzl", "k8s_deploy")

k8s_deploy(
    name = "base",
    gitops = False,
    manifests = ["commonbase.yaml"],
    namespace = "fixednamespace",
)

# first deployment dependent on db
k8s_deploy(
    name = "subsystem_one",
    gitops = False,
    manifests = ["subsystem1.yaml"],
    namespace = "fixednamespace",
    objects = [":base"],
)

k8s_deploy(
    name = "subsystem_two",
    gitops = False,
    manifests = ["subsystem2.yaml"],
    namespace = "fixednamespace",
    objects = [":base"],
)

k8s_deploy(
    name = "final",
    gitops = False,
    manifests = ["join.yaml"],
    namespace = "fixednamespace",
    objects = [
        ":subsystem_one",
        ":subsystem_two",
    ],
)

# test that duplicates in dependency tree are deduplicated
sh_test(
    name = "fork_join_test",
    size = "small",
    srcs = ["fork_join_test.sh"],
    data = [
        "fork_join_expected.yaml",
        ":final.show",
    ],
    deps = ["@bazel_tools//tools/bash/runfiles"],
)
