# Bazel GitOps Rules

![CI](https://github.com/adobe/rules_gitops/workflows/CI/badge.svg?branch=main&event=push)

Bazel GitOps Rules provides tooling to bridge the gap between Bazel (for hermetic, reproducible, container builds) and continuous, git-operation driven, deployments. Users author standard Kubernetes manifests and kustomize overlays for their services. Bazel GitOps Rules handles image push and substitution, applies necessary kustomizations, and handles content addressed substitutions of all object references (configmaps, secrets, etc). Bazel targets are exposed for applying the rendered manifest directly to a Kubernetes cluster, or into version control facilitating deployment via Git operations.

## Features

- **Kustomize Integration**: Full [Kustomize](https://kustomize.io/) capabilities for generating and transforming manifests
- **GitOps Workflow**: Native support for Git-based deployment workflows with automatic PR creation
- **Container Image Management**: Seamless integration with [rules_img](https://github.com/bazel-contrib/rules_img) for building and pushing container images
- **Namespace Deployments**: Support for personal and team namespace deployments
- **Integration Testing**: Built-in utilities for Kubernetes integration test setup
- **Toolchain Support**: Managed kustomize and kubectl toolchains via Bzlmod

## Ruleset Overview

The ruleset is organized into the following modules:

| Module | Description |
|--------|-------------|
| `@rules_gitops//gitops:defs.bzl` | Core deployment rules (`k8s_deploy`, `gitops`) for rendering manifests and managing deployments |
| `@rules_gitops//kustomize:defs.bzl` | Kustomize rules and toolchain for manifest transformation |
| `@rules_gitops//kubectl:defs.bzl` | Kubectl toolchain for cluster interactions |
| `@rules_gitops//testing:defs.bzl` | Integration testing utilities (`k8s_test_setup`, `k8s_test_namespace`) |
| `@rules_gitops//adapters:providers.bzl` | `K8sPushInfo` provider for container image integration |
| `@rules_gitops//adapters:rules_img.bzl` | Adapter for [rules_img](https://github.com/bazel-contrib/rules_img) container images |

## Getting Started

### Installation

Add `rules_gitops` to your `MODULE.bazel`:

```starlark
bazel_dep(name = "rules_gitops", version = "1.0.0")
```

### Toolchain Setup

The ruleset provides managed toolchains for kustomize and kubectl. To use custom versions, configure the extensions in your `MODULE.bazel`:

```starlark
kustomize = use_extension("@rules_gitops//kustomize:defs.bzl", "kustomize")
kustomize.toolchain(
    version = "5.7.1",  # specify your desired version
)

kubectl = use_extension("@rules_gitops//kubectl:defs.bzl", "kubectl")
kubectl.toolchain(
    version = "1.32.2",  # specify your desired version
)
```

<a name="k8s_deploy"></a>
## k8s_deploy

The `k8s_deploy` macro creates rules that produce the `.apply` and `.gitops` targets. `k8s_deploy` takes the files listed in the `manifests`, `patches`, and `configmaps_srcs` attributes and combines (**renders**) them into one YAML file. This happens when you `bazel build` or `bazel run` a target created by the `k8s_deploy`. The file is created at `bazel-bin/path/to/package/name.yaml`. When you run a `.apply` target, it runs `kubectl apply` on this file. When you run a `.gitops` target, it copies this file to the appropriate location in the same or separate repository.

For example, let's look at the [example's k8s_deploy](./examples/helloworld/BUILD.bazel). We can peek at the file containing the rendered K8s manifests:
```bash
cd examples
bazel run //:mynamespace.show
```

### Container Image Integration

`rules_gitops` works with [rules_img](https://github.com/bazel-contrib/rules_img) for container image building and pushing. Use the `k8s_push_info` adapter to connect your images to deployments:

```starlark
load("@rules_gitops//adapters:rules_img.bzl", "k8s_push_info")
load("@rules_gitops//gitops:defs.bzl", "k8s_deploy")
load("@rules_img//img:image.bzl", "image_manifest")
load("@rules_img//img:push.bzl", "image_push")

image_manifest(
    name = "my_image",
    # ... image configuration
)

image_push(
    name = "push",
    image = ":my_image",
    registry = "my-registry.com",
    repository = "my-org/my-image",
)

k8s_push_info(
    name = "k8s_image",
    image = ":my_image",
    push = ":push",
    registry = "my-registry.com",
    repository = "my-org/my-image",
)

k8s_deploy(
    name = "my_deployment",
    cluster = "my-cluster",
    images = [":k8s_image"],
    manifests = ["deployment.yaml"],
    namespace = "{BUILD_USER}",
)
```

### Basic Usage

Once configured, you can use the generated targets:

```bash
# Show rendered manifests (useful for debugging)
bazel run //path/to:my_deployment.show

# Apply manifests to cluster using your kubectl toolchain
bazel run //path/to:my_deployment.apply

# Generate GitOps output
bazel run //path/to:my_deployment.gitops
```

See the [examples/helloworld](./examples/helloworld) directory for a complete working example.

<a name="gitops-workflow"></a>
## GitOps Workflow

The `create_gitops_prs` tool automates the creation of GitOps pull requests as part of your CI pipeline:

The simplified CI pipeline that incorporates GitOps will look like this:
```
[Checkout Code] -> [Bazel Build & Test] -> (if GitOps source branch) -> [Create GitOps PRs]
```

The *Create GitOps PRs* step usually is the last step of a CI pipeline. `rules_gitops` provides the `create_gitops_prs` command line tool that automates the process of creating pull requests.

For the full list of `create_gitops_prs` command line options, run:
```bash
bazel run @rules_gitops//gitops/prer:create_gitops_prs
```

<a name="gitops-and-deployment-supported-git-servers"></a>
### Supported Git Servers

The `--git_server` parameter defines the type of a Git server API to use. The supported Git server types are `github`, `gitlab`, and `bitbucket`.

Depending on the Git server type the `create_gitops_prs` tool will use following command line parameters:

--git_server | Parameter                            | Default
------------ | ------------------------------------ | --------------
| `github`
|            | ***--github_repo_owner***            | ``
|            | ***--github_repo***                  | ``
|            | ***--github_access_token***          | `$GITHUB_TOKEN`
|            | ***--github_enterprise_host***       | ``
| `gitlab`   |
|            | ***--gitlab_host***                  | `https://gitlab.com`
|            | ***--gitlab_repo***                  | ``
|            | ***--gitlab_access_token***          | `$GITLAB_TOKEN`
| `bitbucket`
|            | ***--bitbucket_api_pr_endpoint***    | ``
|            | ***--bitbucket_user***               | `$BITBUCKET_USER`
|            | ***--bitbucket_password***           | `$BITBUCKET_PASSWORD`

<a name="trunk-based-gitops-workflow"></a>
## Trunk Based GitOps Workflow

Let's assume the CI build pipeline described above is running the build for `https://github.com/example/repo.git`. In a trunk based branching model, all feature branches are merged into the `main` branch first. The *Create GitOps PRs* step runs on a `main` branch change. The GitOps deployments source files are located in the same repository under the `/cloud` directory.

The *Create GitOps PRs* pipeline step shell command will look like following:
```bash
GIT_ROOT_DIR=$(git rev-parse --show-toplevel)
GIT_COMMIT_ID=$(git rev-parse HEAD)
GIT_BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)
if [ "${GIT_BRANCH_NAME}" == "master"]; then
    bazel run //gitops/prer:create_gitops_prs -- \
        --workspace $GIT_ROOT_DIR \
        --git_repo https://github.com/example/repo.git \
        --git_mirror $GIT_ROOT_DIR/.git \
        --git_server github \
        --release_branch master \
        --gitops_pr_into master \
        --gitops_pr_title "This is my pull request title" \
        --gitops_pr_body "This is my pull request body message" \
        --branch_name ${GIT_BRANCH_NAME} \
        --git_commit ${GIT_COMMIT_ID} \
fi
```

The `GIT_*` variables describe the current state of the Git repository.

The `--git_repo` parameter defines the remote repository URL. In this case remote repository matches the repository of the working copy. The `--git_mirror` parameter is an optimization used to speed up the target repository clone process using reference repository (see `git clone --reference`). The `--git-server` parameter selects the type of Git server.

The `--release_branch` specifies the value of the ***release_branch_prefix*** attribute of `gitops` targets (see [k8s_deploy](#k8s_deploy)). The `--gitops_pr_into` defines the target branch for newly created pull requests. The `--branch_name` and `--git_commit` are the values used in the pull request commit message.

The `create_gitops_prs` tool will query all `gitops` targets which have set the ***deploy_branch*** attribute (see [k8s_deploy](#k8s_deploy)) and the ***release_branch_prefix*** attribute value that matches the `release_branch` parameter.

The all discovered `gitops` targets are grouped by the value of ***deploy_branch*** attribute. The one deployment branch will accumulate the output of all corresponding `gitops` targets.

For example, we define two deployments: grafana and prometheus. Both deployments share the same namespace. The deployments are grouped by namespace.
```starlark
[
    k8s_deploy(
        name = NAME,
        deploy_branch = NAMESPACE,
        ...
    )
    for NAME, CLUSTER, NAMESPACE in [
        ...
        ("stage-grafana", "stage", "monitoring-stage"),
        ("prod-grafana", "prod", "monitoring-prod"),
    ]
]
[
    k8s_deploy(
        name = NAME,
        deploy_branch = NAMESPACE,
        ...
    )
    for NAME, CLUSTER, NAMESPACE in [
        ...
        ("stage-prometheus", "stage", "monitoring-stage"),
        ("prod-prometheus", "prod", "monitoring-prod"),
    ]
]
```

As a result of the setup above the `create_gitops_prs` tool will open up to 2 potential deployment pull requests:
* from `deploy/monitoring-stage` to `main` including manifests for `stage-grafana` and `stage-prometheus`
* from `deploy/monitoring-prod` to `main` including manifests for `prod-grafana` and `prod-prometheus`

The GitOps pull request is only created (or new commits added) if the `gitops` target changes the state for the target deployment branch. The source pull request will remain open (and keep accumulation GitOps results) until the pull request is merged and source branch is deleted.

The `--stamp` parameter allows for the replacement of certain placeholders, but only when the `gitops` target changes the output's digest compared to the one already saved. The new digest of the unstamped data is also saved with the manifest. The digest is kept in a file in the same location as the YAML file, with a `.digest` extension added to its name. This is helpful when the manifests have volatile information that shouldn't be the only factor causing changes in the target deployment branch.

Here are the placeholders that can be replaced:

| Placeholder      | Replacement                                    |
|------------------|-------------------------------------------------|
| `{{GIT_REVISION}}` | Result of `git rev-parse HEAD`                  |
| `{{UTC_DATE}}`     | Result of `date -u`                             |
| `{{GIT_BRANCH}}`   | The `branch_name` argument given to `create_gitops_prs` |

`--dry_run` parameter can be used to test the tool without creating any pull requests. The tool will print the list of the potential pull requests. It is recommended to run the tool in the dry run mode as a part of the CI test suite to verify that the tool is configured correctly.

<a name="multiple-release-branches-gitops-workflow"></a>
## Multiple Release Branches GitOps Workflow

In the situation when the trunk based branching model in not suitable the `create_gitops_prs` tool supports creating GitOps pull requests before the code is merged to the `main` branch.

Both trunk and release branch workflows can coexist in the same repository.

For example, let's assume the CI build pipeline described above is running the build for `https://github.com/example/repo.git`. In a release branch branching model, features are merged into multiple target release branches. The release brach name convention is `release/team-<YYYYMMDD>`. The *Create GitOps PRs* step runs on release branch changes. GitOps deployments source files are located in the same repository `/cloud` directory in the `main` branch.

The *Create GitOps PRs* pipeline step shell command will look like following:
```bash
GIT_ROOT_DIR=$(git rev-parse --show-toplevel)
GIT_COMMIT_ID=$(git rev-parse HEAD)
GIT_BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)          # => release/team-20200101
RELEASE_BRANCH_SUFFIX=${GIT_BRANCH_NAME#"release/team"}     # => -20200101
RELEASE_BRANCH=${GIT_BRANCH_NAME%${RELEASE_BRANCH_SUFFIX}}  # => release/team
if [ "${RELEASE_BRANCH}" == "release/team"]; then
    bazel run //gitops/prer:create_gitops_prs -- \
        --workspace $GIT_ROOT_DIR \
        --git_repo https://github.com/example/repo.git \
        --git_mirror $GIT_ROOT_DIR/.git \
        --git_server github \
        --release_branch ${RELEASE_BRANCH} \
        --deployment_branch_suffix=${RELEASE_BRANCH_SUFFIX} \
        --gitops_pr_into master \
        --gitops_pr_title "This is my pull request title" \
        --gitops_pr_body "This is my pull request body message" \
        --branch_name ${GIT_BRANCH_NAME} \
        --git_commit ${GIT_COMMIT_ID} \
fi
```

The meaning of the parameters is the same as with [trunk based workflow](#trunk_based_gitops_workflow).
The `--release_branch` parameter takes the value of `release/team`. The additional parameter `--deployment_branch_suffix` will add the release branch suffix to the target deployment branch name.

If we modify the previous example:
```starlark
[
    k8s_deploy(
        name = NAME,
        deploy_branch = NAMESPACE,
        release_branch_prefix = "release/team",  # will only be selected when --release_branch=release/team
        ...
    )
    for NAME, CLUSTER, NAMESPACE in [
        ...
        ("stage-grafana", "stage", "monitoring-stage"),
        ("prod-grafana", "prod", "monitoring-prod"),
    ]
]
[
    k8s_deploy(
        name = NAME,
        deploy_branch = NAMESPACE,
        release_branch_prefix = "release/team",  # will only be selected when --release_branch=release/team
        ...
    )
    for NAME, CLUSTER, NAMESPACE in [
        ...
        ("stage-prometheus", "stage", "monitoring-stage"),
        ("prod-prometheus", "prod", "monitoring-prod"),
    ]
]
```

The result of the setup above the `create_gitops_prs` tool will open up to 2 potential deployment pull requests per release branch. Assuming release branch name is `release/team-20200101`:
* from `deploy/monitoring-stage-20200101` to `master` including manifests for `stage-grafana` and `stage-prometheus`
* from `deploy/monitoring-prod-20200101` to `master` including manifests for `prod-grafana` and `prod-prometheus`


<a name="integration-testing-support"></a>
## Integration Testing Support

**Note:** the Integration testing support has known limitations and should be considered **experimental**. The public API will not abide by semver.

Integration tests are defined in `BUILD.bazel` files like this:
```starlark
k8s_test_setup(
    name = "service_it.setup",
    kubeconfig = "@k8s_test//:kubeconfig",
    objects = [
        "//service:mynamespace",
    ],
)

java_test(
    name = "service_it",
    srcs = [
        "ServiceIT.java",
    ],
    data = [
        ":service_it.setup",
    ],
    jvm_flags = [
        "-Dk8s.setup=$(location :service_it.setup)",
    ],
    # other attributes omitted for brevity
)
```
The test is composed of two rules, a `k8s_test_setup` rule to manage the Kubernetes setup and a `java_test` rule that executes the actual test.

The `k8s_test_setup` rule produces a shell script which creates a temporary namespace (the namespace name is your username followed by five random digits) and creates a kubeconfig file that allows access to this new namespace. Inside the namespace, it creates some objects specified in the `objects` attributes. In the example, there is one target here: `//service:mynamespace`. This target represents a file containing all the Kubernetes object manifests required to run the service.

The output of the `k8s_test_setup` rule (a shell script) is referenced in the `java_test` rule. It's listed under the `data` attribute, which declares the target as a dependency, and is included in the jvm flags in this clause: `$(location :service_it.setup)`. The "location" function is specific to Bazel: given a target, it returns the path to the file produced by that target. In this case, it returns the path to the shell script created by our `k8s_test_setup` rule.

The test code launches the script to perform the test setup. The test code should also monitor the script console output to listen to the pod readiness events.

The `@k8s_test//:kubeconfig` target referenced from `k8s_test_setup` rule serves the purpose of making Kubernetes configuration available in the test sandbox. The `kubeconfig` repository rule in the `WORKSPACE` file will need, at minimum, provide the cluster name.

```starlark
load("//gitops:defs.bzl", "kubeconfig")

kubeconfig(
    name = "k8s_test",
    cluster = "dev",
)
```

## Building & Testing

### Building & Testing GitOps Rules

```bash
bazel test //...
```

### Building & Testing Examples Project

```bash
cd examples/helloworld
bazel test //...
```

## Have a Question

Find the `rules_gitops` contributors in the [#gitops](https://bazelbuild.slack.com/archives/C01SF68MTFS) channel on the [Bazel Slack](https://slack.bazel.build/).

## Contributing

Contributions are welcomed! Read the [Contributing Guide](./.github/CONTRIBUTING.md) for more information.


## Licensing

The contents of [/templating/fasttemplate](./templating/fasttemplate) are licensed under MIT License. See [LICENSE](./templating/fasttemplate/LICENSE) for more information.

All other files are licensed under the Apache V2 License. See [LICENSE](LICENSE) for more information.
