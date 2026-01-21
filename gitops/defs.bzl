# Copyright 2026 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

"""
GitOps rules public interface
"""
load("//gitops/private:k8s_deploy.bzl", _k8s_deploy = "k8s_deploy")
load("//gitops/private:gitops.bzl", _gitops = "gitops")

k8s_deploy = _k8s_deploy
gitops = _gitops
