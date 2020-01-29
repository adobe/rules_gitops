# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

licenses(["notice"])  # Apache 2.0

load("//skylib:k8s.bzl", "k8s_deploy")

k8s_deploy(
    name = "testdeployment",
    cluster = "testcluster",
    deployment_branch = "test1",
    gitops = 1,
    images = {
        "img": "//skylib/kustomize/tests:image",
    },
    namespace = "ci",
    release_branch_prefix = "gitops_test_release_branch",
    visibility = ["//visibility:public"],
)
