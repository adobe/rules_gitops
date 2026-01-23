"""Macro for creating Kubernetes deployment targets with GitOps support."""

load("//gitops/private:gitops.bzl", _gitops = "gitops")
load("//kubectl:defs.bzl", "kubectl_binary")
load("//kustomize:defs.bzl", "kustomization", "show")

def k8s_deploy(
        name,  # name of the rule is important for gitops, since it will become a part of the target manifest file name in /cloud
        cluster = "dev",
        user = "{BUILD_USER}",
        namespace = None,
        configmaps_srcs = None,
        secrets_srcs = None,
        configmaps_renaming = None,  # configmaps renaming policy. Could be None or 'hash'.
        manifests = None,
        name_prefix = None,
        name_suffix = None,
        prefix_suffix_app_labels = False,  # apply kustomize configuration to modify "app" labels in Deployments when name prefix or suffix applied
        patches = None,
        image_name_patches = {},
        image_tag_patches = {},
        substitutions = {},  # dict of template parameter substitutions. CLUSTER and NAMESPACE parameters are added automatically.
        configurations = [],  # additional kustomize configuration files. rules_gitops provides
        common_labels = {},  # list of common labels to apply to all objects see commonLabels kustomize docs
        common_annotations = {},  # list of common annotations to apply to all objects see commonAnnotations kustomize docs
        deps = [],
        deps_aliases = {},
        images = [],
        objects = [],
        gitops = True,  # make sure to use gitops = False to work with individual namespace. This option will be turned False if namespace is '{BUILD_USER}'
        gitops_path = "cloud",
        deployment_branch = None,
        release_branch_prefix = "main",
        start_tag = "{{",
        end_tag = "}}",
        tags = [],
        visibility = None):
    """Generates Kubernetes deployment targets with optional GitOps support.

    This macro creates kustomization targets and kubectl binaries for deploying
    Kubernetes manifests. When gitops is enabled, it also creates a gitops target
    for managing deployments through a GitOps workflow.

    Args:
        name: Name of the rule. Important for gitops since it becomes part of
            the target manifest file name in the gitops_path directory.
        cluster: Target Kubernetes cluster name. Defaults to "dev".
        user: Kubernetes user for authentication. Defaults to "{BUILD_USER}"
            which is substituted at runtime.
        namespace: Target Kubernetes namespace. Required when gitops=True.
            Defaults to "{BUILD_USER}" when gitops=False.
        configmaps_srcs: List of source files for generating ConfigMaps.
        secrets_srcs: List of source files for generating Secrets.
        configmaps_renaming: ConfigMap renaming policy. Can be None (no renaming)
            or "hash" (append content hash to names).
        manifests: List of Kubernetes manifest files. Defaults to all .yaml and
            .yaml.tpl files in the package via glob.
        name_prefix: Prefix to add to all resource names via kustomize.
        name_suffix: Suffix to add to all resource names via kustomize.
        prefix_suffix_app_labels: If True, applies kustomize configuration to
            modify "app" labels in Deployments when name_prefix or name_suffix
            is applied.
        patches: List of kustomize patches to apply to the manifests.
        image_name_patches: Dict mapping original image names to new image names.
        image_tag_patches: Dict mapping image names to new tags.
        substitutions: Dict of template parameter substitutions. CLUSTER and
            NAMESPACE parameters are added automatically and should not be
            included.
        configurations: List of additional kustomize configuration files.
        common_labels: Dict of labels to apply to all Kubernetes objects.
            See kustomize commonLabels documentation.
        common_annotations: Dict of annotations to apply to all Kubernetes
            objects. See kustomize commonAnnotations documentation.
        deps: List of dependency targets.
        deps_aliases: Dict mapping aliases to dependency targets.
        images: List of container image targets to include in the deployment.
        objects: List of additional Kubernetes objects to include.
        gitops: If True, creates a gitops target for GitOps workflow. Set to
            False to work with individual namespaces. Automatically set to
            False if namespace is "{BUILD_USER}".
        gitops_path: Directory path for gitops manifests. Defaults to "cloud".
        deployment_branch: Git branch for deployments. If None, uses default.
        release_branch_prefix: Prefix for release branches. Defaults to "main".
        start_tag: Opening delimiter for template substitutions. Defaults to "{{".
        end_tag: Closing delimiter for template substitutions. Defaults to "}}".
        tags: List of tags to apply to all generated targets.
        visibility: Visibility specification for generated targets.
    """

    if not manifests:
        manifests = native.glob(["*.yaml", "*.yaml.tpl"])
    if prefix_suffix_app_labels:
        configurations = configurations + [
            "//gitops:nameprefix_deployment_labels_config.yaml",
            "//gitops:namesuffix_deployment_labels_config.yaml",
        ]
    for reservedname in ["CLUSTER", "NAMESPACE"]:
        if substitutions.get(reservedname):
            fail("do not put %s in substitutions parameter of k8s_deploy. It will be added autimatically" % reservedname)
    substitutions = dict(substitutions)
    substitutions["CLUSTER"] = cluster

    # NAMESPACE substitution is deferred until test_setup/kubectl/gitops
    if namespace == "{BUILD_USER}":
        gitops = False

    if not gitops:
        # Mynamespace option
        if not namespace:
            namespace = "{BUILD_USER}"
        kustomization(
            name = name,
            namespace = namespace,
            configmaps_srcs = configmaps_srcs,
            secrets_srcs = secrets_srcs,
            # disable_name_suffix_hash is renamed to configmaps_renaming in recent Kustomize
            disable_name_suffix_hash = (configmaps_renaming != "hash"),
            images = images,
            manifests = manifests,
            substitutions = substitutions,
            deps = deps,
            deps_aliases = deps_aliases,
            start_tag = start_tag,
            end_tag = end_tag,
            name_prefix = name_prefix,
            name_suffix = name_suffix,
            configurations = configurations,
            common_labels = common_labels,
            common_annotations = common_annotations,
            patches = patches,
            objects = objects,
            image_name_patches = image_name_patches,
            image_tag_patches = image_tag_patches,
            tags = tags,
            visibility = visibility,
        )
        kubectl_binary(
            name = name + ".apply",
            srcs = [name],
            cluster = cluster,
            user = user,
            namespace = namespace,
            tags = tags,
            visibility = visibility,
        )
        kubectl_binary(
            name = name + ".delete",
            srcs = [name],
            command = "delete",
            cluster = cluster,
            push = False,
            user = user,
            namespace = namespace,
            tags = tags,
            visibility = visibility,
        )
        show(
            name = name + ".show",
            namespace = namespace,
            src = name,
            tags = tags,
            visibility = visibility,
        )
    else:
        # gitops
        if not namespace:
            fail("namespace must be defined for gitops k8s_deploy")
        kustomization(
            name = name,
            namespace = namespace,
            configmaps_srcs = configmaps_srcs,
            secrets_srcs = secrets_srcs,
            # disable_name_suffix_hash is renamed to configmaps_renaming in recent Kustomize
            disable_name_suffix_hash = (configmaps_renaming != "hash"),
            images = images,
            manifests = manifests,
            visibility = visibility,
            substitutions = substitutions,
            deps = deps,
            deps_aliases = deps_aliases,
            start_tag = start_tag,
            end_tag = end_tag,
            name_prefix = name_prefix,
            name_suffix = name_suffix,
            configurations = configurations,
            common_labels = common_labels,
            common_annotations = common_annotations,
            patches = patches,
            image_name_patches = image_name_patches,
            image_tag_patches = image_tag_patches,
            tags = tags,
        )
        kubectl_binary(
            name = name + ".apply",
            srcs = [name],
            cluster = cluster,
            user = user,
            namespace = namespace,
            tags = tags,
            visibility = visibility,
        )
        _gitops(
            name = name + ".gitops",
            srcs = [name],
            cluster = cluster,
            namespace = namespace,
            gitops_path = gitops_path,
            strip_prefixes = [
                namespace + "-",
                cluster + "-",
            ],
            deployment_branch = deployment_branch,
            release_branch_prefix = release_branch_prefix,
            tags = tags,
            visibility = ["//visibility:public"],
        )
        show(
            name = name + ".show",
            src = name,
            namespace = namespace,
            tags = tags,
            visibility = visibility,
        )
