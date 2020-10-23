# Bazel GitOps Rules

![CI](https://github.com/adobe/rules_gitops/workflows/CI/badge.svg?branch=master&event=push)

Bazel GitOps Rules provides tooling to bridge the gap between Bazel (for hermetic, reproducible, container builds) and continuous, git-operation driven, deployments. Users author standard kubernetes manifests and kustomize overlays for their services. Bazel GitOps Rules handles image push and substitution, applies necessary kustomizations, and handles content addressed substitutions of all object references (configmaps, secrets, etc). Bazel targets are exposed for applying the rendered manifest directly to a Kubernetes cluster, or into version control facilitating deployment via Git operations.

Bazel GitOps Rules is an alternative to [rules_k8s](https://github.com/bazelbuild/rules_k8s). The main differences are:

* Utilizes and integrates the full set of [Kustomize](https://kustomize.io/) capabilities for generating manifests.
* Implements GitOps target.
* Supports personal namespace deployments.
* Provides integration test setup utility.
* Speeds up deployments iterations:
  * The results manifests are rendered without pushing containers.
  * Pushes all the images in parallel.


## Rules

* [k8s_deploy](#k8s_deploy)
* [k8s_test_setup](#k8s_test_setup)


## Setup

Add the following to your `WORKSPACE` file to add the necessary external dependencies:

<!--
# generate the WORKSPACE snippet:

rev=$(git rev-parse HEAD) && sha265=$(curl -Ls https://github.com/adobe/rules_gitops/archive/${rev}.zip | shasum -a 256 - | cut -d ' ' -f1) && cat <<EOF
# copy/paste following snippet into README.md
rules_gitops_version = "${rev}"

http_archive(
    name = "com_adobe_rules_gitops",
    sha256 = "${sha265}",
    strip_prefix = "rules_gitops-%s" % rules_gitops_version,
    urls = ["https://github.com/adobe/rules_gitops/archive/%s.zip" % rules_gitops_version],
)
EOF
-->

```python
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

rules_gitops_version = "8d9416a36904c537da550c95dc7211406b431db9"

http_archive(
    name = "com_adobe_rules_gitops",
    sha256 = "25601ed932bab631e7004731cf81a40bd00c9a34b87c7de35f6bc905c37ef30d",
    strip_prefix = "rules_gitops-%s" % rules_gitops_version,
    urls = ["https://github.com/adobe/rules_gitops/archive/%s.zip" % rules_gitops_version],
)

load("@com_adobe_rules_gitops//gitops:deps.bzl", "rules_gitops_dependencies")

rules_gitops_dependencies()

load("@com_adobe_rules_gitops//gitops:repositories.bzl", "rules_gitops_repositories")

rules_gitops_repositories()

```


<a name="k8s_deploy"></a>
## k8s_deploy

The `k8s_deploy` creates rules that produce the `.apply` and `.gitops` targets `k8s_deploy` is defined in [k8s.bzl](./skylib/k8s.bzl). `k8s_deploy` takes the files listed in the `manifests`, `patches`, and `configmaps_srcs` attributes and combines (**renders**) them into one  YAML file. This happens when you `bazel build` or `bazel run` a target created by the `k8s_deploy`. The file is created at `bazel-bin/path/to/package/name.yaml`. When you run a `.apply` target, it runs `kubectl apply` on this file. When you run a `.gitops` target, it copies this file to
the appropriate location in the same os separate repository.

For example, let's look at the [example's k8s_deploy](./examples/helloworld/BUILD). We can peek at the file containing the rendered K8s manifests:
```bash
cd examples
bazel run //helloworld:mynamespace.show
```
When you run `bazel run ///helloworld:mynamespace.apply`, it applies this file into your personal (`{BUILD_USER}`) namespace. Viewing the rendered files with `.show` can be useful for debugging issues with invalid or misconfigured manifests.

| Parameter                 | Default        | Description
| ------------------------- | -------------- | -----------
| ***cluster***             | `None`         | The name of the cluster in which these manifests will be applied.
| ***namespace***           | `None`         | The target namespace to assign to all manifests. Any namespace value in the source manifests will be replaced or added if not specified.
| ***user***                | `{BUILD_USER}` | The user passed to kubectl in .apply rule. Must exist in users ~/.kube/config
| ***configmaps_srcs***     | `None`         | A list of files (of any type) that will be combined into configmaps. See [Generating Configmaps](#generating-configmaps).
| ***configmaps_renaming*** | `None`         | Configmaps/Secrets renaming policy. Could be None or 'hash'. 'hash' renaming policy is used to add a unique suffix to the generated configmap or secret name. All references to the configmap or secret in other manifests will be replaced with the generated name.
| ***secrets_srcs***        | `None`         | A list of files (of any type) that will be combined into a secret similar to configmaps.
| ***manifests***           | `glob(['*.yaml','*.yaml.tpl'])` | A list of base manifests. See [Base Manifests and Overlays](#base-manifests-and-overlays).
| ***name_prefix***         | `None`         | Adds prefix to the names of all resources defined in manifests.
| ***name_suffix***         | `None`         | Adds suffix to the names of all resources defined in manifests.
| ***patches***             | `None`         | A list of patch files to overlay the base manifests. See [Base Manifests and Overlays](#base-manifests-and-overlays).
| ***substitutions***       | `None`         | Does parameter substitution in all the manifests (including configmaps). This should generally be limited to "CLUSTER" and "NAMESPACE" only. Any other replacements should be done with overlays.
| ***configurations***      | `[]`           | A list of files with [kustomize configurations](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/transformerconfigs/README.md).
| ***prefix_suffix_app_labels*** | `False`   | Add the bundled configuration file allowing adding suffix and prefix to labels `app` and `app.kubernetes.io/name` and respective selector in Deployment.
| ***common_labels***       | `{}`           | A map of labels that should be added to all objects and object templates.
| ***common_annotations***  | `{}`           | A map of annotations that should be added to all objects and object templates.
| ***start_tag***           | `"{{"`         | The character start sequence used for substitutions.
| ***end_tag***             | `"}}"`         | The character end sequence used for substitutions.
| ***deps***                | `[]`           | A list of dependencies used to drive `k8s_deploy` functionality (i.e. `deps_aliases`).
| ***deps_aliases***        | `{}`           | A dict of labels of file dependencies. File dependency contents are available for template expansion in manifests as `{{imports.<label>}}`. Each dependency in this dictionary should be present in the `deps` attribute.
| ***objects***             | `[]`           | A list of other instances of `k8s_deploy` that this one depends on. See [Adding Dependencies](#adding-dependencies).
| ***images***              | `{}`           | A dict of labels of Docker images. See [Injecting Docker Images](#injecting-docker-images).
| ***image_digest_tag***    | `False`        | A flag for whether or not to tag the image with the container digest.
| ***image_registry***      | `docker.io`    | The registry to push images to. 
| ***image_repository***    | `None`         | The repository to push images to. By default, this is generated from the current package path.
| ***image_repository_prefix*** | `None`     | Add a prefix to the image_repository. Can be used to upload the images in
| ***release_branch_prefix*** | `master`     | A git branch name/prefix. Automatically run GitOps while building this branch. See [GitOps and Deployment](#gitops_and_deployment).
| ***deployment_branch***   | `None`         | Automatic GitOps output will appear in a branch and PR with this name. See [GitOps and Deployment](#gitops_and_deployment).
| ***visibility***          | [Default_visibility](https://docs.bazel.build/versions/master/be/functions.html#package.default_visibility) | Changes the visibility of all rules generated by this macro. See [Bazel docs on visibility](https://docs.bazel.build/versions/master/be/common-definitions.html#common-attributes).

<a name="base-manifests-and-overlays"></a>
### Base Manifests and Overlays

The manifests listed in the _manifests_ attribute are the base manifests used by the deployment. This is where the important manifests like Deployments, Services, etc. are listed.

The base manifests will be modified by most of the other `k8s_deploy` attributes like `substitutions` and `images`. Additionally, they can be modified to configure them different clusters/namespaces/etc. using **overlays**.

To demonstrate, let's go over hypothetical multi cluster deployment.

Here is the fragment of the `k8s_deploy` rule that is responsible for generating manifest variants per CLOUD, CLUSTER, and NAMESPACE :
```python
k8s_deploy(
    ...
    manifests = glob([                 # (1)
      "manifests/*.yaml",
      "manifests/%s/*.yaml" % (CLOUD),
    ]),
    patches = glob([                   # (2)
      "overlays/*.yaml",
      "overlays/%s/*.yaml" % (CLOUD),
      "overlays/%s/%s/*.yaml" % (CLOUD, NAMESPACE),
      "overlays/%s/%s/%s/*.yaml" % (CLOUD, NAMESPACE, CLUSTER),
    ]),
    ...
)
```
The manifests list `(1)` combines common base manifests and `CLOUD` specific manifests.
```
manifests
├── aws
│   └── pvc.yaml
├── onprem
│   ├── pv.yaml
│   └── pvc.yaml
├── deployment.yaml
├── ingress.yaml
└── service.yaml
```
Here we see that `aws` and `onprem` clouds have different persistence configurations `aws/pvc.yaml` and `onprem/pvc.yaml`.

The patches list `(2)` requires more granular configuration that introduces 3 levels of customization: CLOUND, NAMESPACE, and CLUSTER. Each manifest fragment in the overlays subtree applied as [strategic merge patch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md) update operation.
```
overlays
├── aws
│   ├── deployment.yaml
│   ├── prod
│   │   ├── deployment.yaml
│   │   └── us-east-1
│   │       └── deployment.yaml
│   └── uat
│       └── deployment.yaml
└── onprem
    ├── prod
    │   ├── deployment.yaml
    │   └── us-east
    │       └── deployment.yaml
    └── uat
        └── deployment.yaml
```
That looks like a lot. But lets try to decode what is happening here:

1. `aws/deployment.yaml` adds persistent volume reference specific to all AWS deployments.
1. `aws/prod/deployment.yaml` modifies main container CPU and memory requirements in production configurations.
1. `aws/prod/us-east-1/deployment.yaml` adds monitoring sidecar.


<a name="generating-configmaps"></a>
### Generating Configmaps

Configmaps are a special case of manifests. They can be rendered from a collection of files of any kind (.yaml, .properties, .xml, .sh, whatever). Let's use hypothetical Grafana deployment as an example:

```python
[k8s_deploy(
        name = NAME,
        cluster = CLUSTER,
        configmaps_srcs = glob([                 # (1)
            "configmaps/%s/**/*" % CLUSTER
        ]),
        configmaps_renaming = 'hash',            # (2)

        ...
    )
    for NAME, CLUSTER, NAMESPACE in [
        ("mynamespace", "dev", "{BUILD_USER}"),  # (3)
        ("prod-grafana", "prod", "prod"),        # (4)
    ]
]
```
Here we generate two `k8s_deploy` targets, one for `mynamespace` `(3)`, another for production deployment `(4)`.

The directory structure of `configmaps` looks like this:
```
grafana
└── configmaps
    ├── dev
    │   └── grafana
    │       └── ldap.toml
    └── prod
        └── grafana
            └── ldap.toml
```
The `configmaps_srcs` parameter `(1)` will get resolved into the patterns `configmaps/dev/**/*` and `configmaps/prod/**/*`. The result of rendering the manifests `bazel run //grafana:prod-grafana.show` will have following manifest fragment:

```yaml
apiVersion: v1
data:
  ldap.toml: |
    [[servers]]
    ...
kind: ConfigMap
metadata:
  name: grafana-k75h878g4f
  namespace: ops-prod
```
The name of directory on the first level of glob patten `grafana` become the configmap name. The `ldap.toml` file on the next level were embedded into the configmap.

In this example, the configmap renaming policy `(2)` is set to `hash`, so the configmap's name appears as `grafana-k75h878g4f`. (If the renaming policy was `None`, the configmap's name would remain as `grafana`.) All the references to the `grafana` configmap in other manifests are replaced with the generated name:

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      volumes:
      ...
      - configMap:
          items:
          - key: ldap.toml
            path: ldap.toml
          name: grafana-k75h878g4f
        name: grafana-ldap
```

<a name="injecting-docker-images"></a>
### Injecting Docker Images

Third-party Docker images can be referenced directly in K8s manifests, but for most apps, we need to run our own images. The images are built in the Bazel build pipeline using [rules_docker](https://github.com/bazelbuild/rules_docker). For example, the `java_image` rule creates an image of a Java application from Java source code, dependencies, and configuration.

Here's a (very contrived) example of how this ties in with `k8s_deploy`. Here's the BUILD file:
```python
java_image(
    name = "some_java_image",
    srcs = glob(["*.java"]),
    ...
)
k8s_deploy(
    name = "example",
    manifests = ["my_pod.yaml"],
    images = {
        "my_pod_image": ":some_java_image",
    }
)
```
And here's "my_pod.yaml":
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my_pod
spec:
  containers:
    - name: java_container
      image: my_pod_image
```
When we `bazel build` the example, the rendered manifest will look something like this:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my_pod
spec:
  containers:
    - name: java_container
      image: registry.example.com/examples/image@sha256:c94d75d68f4c1b436f545729bbce82774fda07
```
That URL points to the "some_java_image" in the private Docker registry. The image is uploaded to the registry before any `.apply` or `.gitops` target is executed.

As with the rest of the dependency graph, Bazel understands the dependencies `k8s_deploy` has on the
Docker image and the files in the image. So for example, here's what will happen if someone makes a change to one of the Java files in "some_java_image" and then runs `bazel run //:example.apply`:
1. The "some_java_image" will be rebuilt with the new code and uploaded to the registry
1. A new "my_pod" manifest will be rendered using the new image
1. The new "my_pod" will be deployed


<a name="adding-dependencies"></a>
### Adding Dependencies

Many instances of `k8s_deploy` include an `objects` attribute that references other instances of
`k8s_deploy`. When chained this way, running the `.apply` will also apply any dependencies as well.

For example, to add dependency to the example [helloworld deployment](./examples/helloworld/BUILD):
```python
k8s_deploy(
    name = "mynamespace",
    objects = [
        "//other:mynamespace",
    ],
    ...
)
```
When you run `bazel run //helloworld:mynamespace.apply`, it'll deploy a _helloword_ and _other_  service instance into your namespace.

Please note that the `objects` attribute is ignored by `.gitops` targets.


<a name="gitops_and_deployment"></a>
### GitOps and Deployment

<!-- TODO(KZ): document gitops and deployment -->


<a name="k8s_test_setup"></a>
## k8s_test_setup

Integration tests are defined in `BUILD` files like this:
```python
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

The test code launches the script to perform the test setup. The tes code should also monitor the script console output to listen to the pod readiness events.


## Building & Testing

### Building & Testing GitOps Rules

```bash
bazel test //...
```

### Building & Testing Examples Project

```bash
cd examples
bazel test //...
```


## Contributing

Contributions are welcomed! Read the [Contributing Guide](./.github/CONTRIBUTING.md) for more information.


## Adopters
Here's a (non-exhaustive) list of companies that use `rules_gitops` in production. Don't see yours? [You can add it in a PR!](https://github.com/adobe/rules_gitops/edit/master/README.md)
  * [Adobe (Advertising Cloud)](https://www.adobe.com/advertising/adobe-advertising-cloud.html)


## Licensing

The contents of third party dependencies in [/vendor](./vendor) folder are covered by their repositories' respective licenses.

The contents of [/templating/fasttemplate](./templating/fasttemplate) are licensed under MIT License. See [LICENSE](./templating/fasttemplate/LICENSE) for more information.

All other files are licensed under the Apache V2 License. See [LICENSE](LICENSE) for more information.
