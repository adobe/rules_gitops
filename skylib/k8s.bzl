# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

load(
    "@io_bazel_rules_docker//skylib:path.bzl",
    _get_runfile_path = "runfile",
)
load(
    "//skylib/kustomize:kustomize.bzl",
    "KustomizeInfo",
    "imagePushStatements",
    "kubectl",
    "kustomize",
    kustomize_gitops = "gitops",
)
load("//skylib:push.bzl", "k8s_container_push")

def _runfiles(ctx, f):
    return "${RUNFILES}/%s" % _get_runfile_path(ctx, f)

def _python_runfiles(ctx, f):
    return "PYTHON_RUNFILES=${RUNFILES} %s" % _runfiles(ctx, f)

def _show_impl(ctx):
    script_content = "#!/bin/bash\nset -e\n"

    kustomize_outputs = []
    script_template = "{template_engine} --template={infile} --variable=NAMESPACE={namespace} --stamp_info_file={info_file}\n"
    for dep in ctx.attr.src.files.to_list():
        kustomize_outputs.append(script_template.format(
            infile = dep.short_path,
            template_engine = ctx.executable._template_engine.short_path,
            namespace = ctx.attr.namespace,
            info_file = ctx.file._info_file.short_path,
        ))

    # ensure kustomize outputs are separated by '---' delimiters
    script_content += "echo '---'\n".join(kustomize_outputs)

    ctx.actions.write(ctx.outputs.executable, script_content, is_executable = True)
    return [
        DefaultInfo(runfiles = ctx.runfiles(files = [ctx.executable._template_engine, ctx.file._info_file] + ctx.files.src)),
    ]

show = rule(
    implementation = _show_impl,
    attrs = {
        "src": attr.label(
            doc = "Input file.",
            mandatory = True,
        ),
        "namespace": attr.string(
            doc = "kubernetes namespace.",
            mandatory = True,
        ),
        "_info_file": attr.label(
            default = Label("//skylib:more_stable_status.txt"),
            allow_single_file = True,
        ),
        "_template_engine": attr.label(
            default = Label("//templating:fast_template_engine"),
            executable = True,
            cfg = "host",
        ),
    },
    executable = True,
)

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
        patches = None,
        substitutions = {},  # dict of template parameter substitutions. CLUSTER and NAMESPACE parameters are added automatically.
        deps = [],
        deps_aliases = {},
        images = {},
        image_chroot = None,  # DEPRECATED. ignored now. If default repo path is not working for you, use image_repository to change it
        image_registry = "docker.io",  # registry to push container to. jenkins will need an access configured for gitops to work. Ignored for mynamespace.
        image_repository = None,  # repository (registry path) to push container to. Generated from the image bazel path if empty.
        image_repository_prefix = None,  # Mutually exclusive with 'image_repository'. Add a prefix to the repository name generated from the image bazel path
        objects = [],
        gitops = True,  # make sure to use gitops = False to work with individual namespace. This option will be turned False if namespace is '{BUILD_USER}'
        deployment_branch = None,
        release_branch_prefix = "master",
        flatten_manifest_directories = False,
        start_tag = "{{",
        end_tag = "}}",
        visibility = None):
    """ k8s_deploy
    """

    if not manifests:
        manifests = native.glob(["*.yaml", "*.yaml.tpl"])
    for reservedname in ["CLUSTER", "NAMESPACE"]:
        if substitutions.get(reservedname):
            fail("do not put %s in substitutions parameter of k8s_deploy. It will be added autimatically" % reservedname)
    substitutions = dict(substitutions)
    substitutions["CLUSTER"] = cluster

    # NAMESPACE substitution is deferred until test_setup/kubectl/gitops
    if namespace == "{BUILD_USER}":
        gitops = False
    if image_chroot:
        print("image_chroot parameter of k8s_deploy rule in %s is ignored now. If default repo path is not working for you, use image_repository to change it." % native.package_name())
    if not gitops:
        # Mynamespace option
        if not namespace:
            namespace = "{BUILD_USER}"
        images_v = []
        for imgname in images:
            img = images[imgname]
            images_v.append(imgname + "_mynamespace_push")
            if not native.existing_rule(imgname + "_mynamespace_push"):
                k8s_container_push(
                    name = imgname + "_mynamespace_push",
                    image = img,
                    legacy_image_name = imgname,
                    registry = image_registry,
                    repository = image_repository,
                    repository_prefix = "{BUILD_USER}",
                )
        kustomize(
            name = name,
            namespace = namespace,
            configmaps_srcs = configmaps_srcs,
            secrets_srcs = secrets_srcs,
            # disable_name_suffix_hash is renamed to configmaps_renaming in recent Kustomize
            disable_name_suffix_hash = (configmaps_renaming != "hash"),
            images = images_v,
            manifests = manifests,
            substitutions = substitutions,
            deps = deps,
            deps_aliases = deps_aliases,
            start_tag = start_tag,
            end_tag = end_tag,
            name_prefix = name_prefix,
            name_suffix = name_suffix,
            patches = patches,
            objects = objects,
            visibility = visibility,
        )
        kubectl(
            name = name + ".apply",
            srcs = [name],
            cluster = cluster,
            user = user,
            namespace = namespace,
            visibility = visibility,
        )
        kubectl(
            name = name + ".delete",
            srcs = [name],
            command = "delete",
            cluster = cluster,
            user = user,
            namespace = namespace,
            visibility = visibility,
        )
        show(
            name = name + ".show",
            namespace = namespace,
            src = name,
            visibility = visibility,
        )
    else:
        # gitops
        if objects:
            print("Warning: objects parameter of k8s_deploy should not be used for gitops in %s. make sure dependencies are processed independently." % native.package_name())
        if not namespace:
            fail("namespace must be defined for gitops k8s_deploy")
        images_v = []
        for imgname in images:
            img = images[imgname]
            images_v.append(imgname + "_push")
            if not native.existing_rule(imgname + "_push"):
                k8s_container_push(
                    name = imgname + "_push",
                    image = img,
                    legacy_image_name = imgname,
                    registry = image_registry,
                    repository = image_repository,
                    repository_prefix = image_repository_prefix,
                )
        kustomize(
            name = name,
            namespace = namespace,
            configmaps_srcs = configmaps_srcs,
            secrets_srcs = secrets_srcs,
            # disable_name_suffix_hash is renamed to configmaps_renaming in recent Kustomize
            disable_name_suffix_hash = (configmaps_renaming != "hash"),
            images = images_v,
            manifests = manifests,
            visibility = visibility,
            substitutions = substitutions,
            deps = deps,
            deps_aliases = deps_aliases,
            start_tag = start_tag,
            end_tag = end_tag,
            name_prefix = name_prefix,
            name_suffix = name_suffix,
            patches = patches,
        )
        kubectl(
            name = name + ".apply",
            srcs = [name],
            cluster = cluster,
            user = user,
            namespace = namespace,
            visibility = visibility,
        )
        kustomize_gitops(
            name = name + ".gitops",
            srcs = [name],
            cluster = cluster,
            namespace = namespace,
            strip_prefixes = [
                namespace + "-",
                cluster + "-",
            ],
            deployment_branch = deployment_branch,
            release_branch_prefix = release_branch_prefix,
            visibility = ["//visibility:public"],
        )
        show(
            name = name + ".show",
            src = name,
            namespace = namespace,
            visibility = visibility,
        )

# kubectl template
def _kubectl_config(repository_ctx, args):
    kubectl = repository_ctx.path("kubectl")
    kubeconfig_yaml = repository_ctx.path("kubeconfig")
    exec_result = repository_ctx.execute(
        [kubectl, "--kubeconfig", kubeconfig_yaml, "config"] + args,
        environment = {
            # prevent kubectl config to stumble on shared .kube/config.lock file
            "HOME": str(repository_ctx.path(".")),
        },
        quiet = True,
    )
    if exec_result.return_code != 0:
        fail("Error executing kubectl config %s" % " ".join(args))

def _kubeconfig_impl(repository_ctx):
    """Find local kubernetes certificates"""

    # find and symlink kubectl
    kubectl = repository_ctx.which("kubectl")
    if not kubectl:
        fail("Unable to find kubectl executable. PATH=%s" % repository_ctx.path)
    repository_ctx.symlink(kubectl, "kubectl")

    # TODO: figure out how to use BUILD_USER
    if "USER" in repository_ctx.os.environ:
        user = repository_ctx.os.environ["USER"]
    else:
        exec_result = repository_ctx.execute(["whoami"])
        if exec_result.return_code != 0:
            fail("Error detecting current user")
        user = exec_result.stdout.rstrip()
    token = None
    ca_crt = None
    kubecert_cert = None
    kubecert_key = None
    server = repository_ctx.attr.server

    # check service account first
    serviceaccount = repository_ctx.path("/var/run/secrets/kubernetes.io/serviceaccount")
    if serviceaccount.exists:
        ca_crt = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
        token_file = serviceaccount.get_child("token")
        if token_file.exists:
            exec_result = repository_ctx.execute(["cat", token_file.realpath])
            if exec_result.return_code != 0:
                fail("Error reading user token")
            token = exec_result.stdout.rstrip()

        # use master url from the environemnt
        if "KUBERNETES_SERVICE_HOST" in repository_ctx.os.environ:
            server = "https://%s:%s" % (
                repository_ctx.os.environ["KUBERNETES_SERVICE_HOST"],
                repository_ctx.os.environ["KUBERNETES_SERVICE_PORT"],
            )
        else:
            # fall back to the default
            server = "https://kubernetes.default"
        print("Using in cluster configuration. Kubernetes master is running at %s" % server)
    else:
        home = repository_ctx.path(repository_ctx.os.environ["HOME"])
        certs = home.get_child(".kube").get_child("certs")
        ca_crt = certs.get_child("ca.crt").realpath
        kubecert_cert = certs.get_child("kubecert.cert")
        kubecert_key = certs.get_child("kubecert.key")

    # config set-cluster {cluster} \
    #     --certificate-authority=... \
    #     --server=https://dev3.k8s.tubemogul.info:443 \
    #     --embed-certs",
    _kubectl_config(repository_ctx, [
        "set-cluster",
        repository_ctx.attr.cluster,
        "--server",
        server,
        "--certificate-authority",
        ca_crt,
    ])

    # config set-credentials {user} --token=...",
    if token:
        _kubectl_config(repository_ctx, [
            "set-credentials",
            user,
            "--token",
            token,
        ])

    # config set-credentials {user} --client-certificate=... --embed-certs",
    if kubecert_cert and kubecert_cert.exists:
        _kubectl_config(repository_ctx, [
            "set-credentials",
            user,
            "--client-certificate",
            kubecert_cert.realpath,
        ])

    # config set-credentials {user} --client-key=... --embed-certs",
    if kubecert_key and kubecert_key.exists:
        _kubectl_config(repository_ctx, [
            "set-credentials",
            user,
            "--client-key",
            kubecert_key.realpath,
        ])

    # export repostory contents
    repository_ctx.file("BUILD", """exports_files(["kubeconfig", "kubectl"])""", False)

    return {
        "cluster": repository_ctx.attr.cluster,
        "server": repository_ctx.attr.server,
    }

kubeconfig = repository_rule(
    attrs = {
        "cluster": attr.string(),
        "server": attr.string(),
    },
    environ = [
        "HOME",
        "USER",
        "KUBERNETES_SERVICE_HOST",
        "KUBERNETES_SERVICE_PORT",
    ],
    local = True,
    implementation = _kubeconfig_impl,
)

def _stamp(ctx, string, output):
    stamps = [ctx.file._info_file]
    stamp_args = [
        "--stamp-info-file=%s" % sf.path
        for sf in stamps
    ]
    ctx.actions.run(
        executable = ctx.executable._stamper,
        arguments = [
            "--format=%s" % string,
            "--output=%s" % output.path,
        ] + stamp_args,
        inputs = [ctx.executable._stamper] + stamps,
        outputs = [output],
        mnemonic = "Stamp",
    )

def _k8s_cmd_impl(ctx):
    files = [ctx.executable._stamper, ctx.file.kubectl, ctx.file.kubeconfig]

    # replace placeholders in the parameter value
    # see: https://github.com/bazelbuild/rules_docker#stamping
    command_arg = ctx.expand_make_variables("command", ctx.attr.command, {})
    if "{" in ctx.attr.command:
        command_file = ctx.actions.declare_file(ctx.label.name + ".command")
        _stamp(ctx, ctx.attr.command, command_file)
        command_arg = "source %s" % _runfiles(ctx, command_file)
        files += [command_file]

    ctx.actions.expand_template(
        template = ctx.file._template,
        substitutions = {
            "%{kubeconfig}": ctx.file.kubeconfig.path,
            "%{kubectl}": ctx.file.kubectl.path,
            "%{statements}": command_arg,
        },
        output = ctx.outputs.executable,
    )
    return [
        DefaultInfo(runfiles = ctx.runfiles(files = files)),
    ]

_k8s_cmd = rule(
    attrs = {
        "command": attr.string(),
        "kubeconfig": attr.label(
            allow_single_file = True,
        ),
        "kubectl": attr.label(
            cfg = "host",
            executable = True,
            allow_single_file = True,
        ),
        "_info_file": attr.label(
            default = Label("//skylib:more_stable_status.txt"),
            allow_single_file = True,
        ),
        "_stamper": attr.label(
            default = Label("@io_bazel_rules_k8s//k8s:stamper"),
            cfg = "host",
            executable = True,
            allow_files = True,
        ),
        "_template": attr.label(
            default = Label("//skylib:k8s_cmd.sh.tpl"),
            allow_single_file = True,
        ),
    },
    executable = True,
    implementation = _k8s_cmd_impl,
)

def _k8s_test_namespace_impl(ctx):
    files = []  # runfiles list

    # add files referenced by rule attributes to runfiles
    files = [ctx.file.kubectl, ctx.file.kubeconfig]

    # create namespace reservation script
    namespace_create = ctx.actions.declare_file(ctx.label.name + ".create")
    ctx.actions.expand_template(
        template = ctx.file._namespace_template,
        substitutions = {
            "%{kubeconfig}": ctx.file.kubeconfig.path,
            "%{kubectl}": ctx.file.kubectl.path,
        },
        output = namespace_create,
        is_executable = True,
    )
    files += [namespace_create]

    return [DefaultInfo(
        executable = namespace_create,
        runfiles = ctx.runfiles(files = files),
    )]

k8s_test_namespace = rule(
    attrs = {
        "kubeconfig": attr.label(
            allow_single_file = True,
        ),
        "kubectl": attr.label(
            cfg = "host",
            executable = True,
            allow_single_file = True,
        ),
        "_namespace_template": attr.label(
            default = Label("//skylib:k8s_test_namespace.sh.tpl"),
            allow_single_file = True,
        ),
    },
    executable = True,
    implementation = _k8s_test_namespace_impl,
)

def _k8s_test_setup_impl(ctx):
    files = []  # runfiles list
    transitive = []
    commands = []  # the list of commands to execute

    # add files referenced by rule attributes to runfiles
    files = [ctx.executable._stamper, ctx.file.kubectl, ctx.file.kubeconfig, ctx.executable._kustomize, ctx.executable._it_sidecar, ctx.executable._it_manifest_filter]
    files += ctx.files._set_namespace

    push_statements, files, pushes_runfiles = imagePushStatements(ctx, [o for o in ctx.attr.objects if KustomizeInfo in o], files)

    # execute all objects targets
    for obj in ctx.attr.objects:
        if obj.files_to_run.executable:
            # add object' targets and excutables to runfiles
            files += [obj.files_to_run.executable]
            transitive.append(obj.default_runfiles.files)

            # add object' execution command
            commands += [_runfiles(ctx, obj.files_to_run.executable) + " | ${SET_NAMESPACE} $NAMESPACE | ${IT_MANIFEST_FILTER} | ${KUBECTL} apply -f -"]
        else:
            files += obj.files.to_list()
            commands += [ctx.executable._template_engine.short_path + " --template=" + filename.short_path + " --variable=NAMESPACE=${NAMESPACE} | ${SET_NAMESPACE} $NAMESPACE | ${IT_MANIFEST_FILTER} | ${KUBECTL} apply -f -" for filename in obj.files.to_list()]

    files += [ctx.executable._template_engine]

    # create namespace script
    ctx.actions.expand_template(
        template = ctx.file._namespace_template,
        substitutions = {
            "%{it_sidecar}": ctx.executable._it_sidecar.short_path,
            "%{kubeconfig}": ctx.file.kubeconfig.path,
            "%{kubectl}": ctx.file.kubectl.path,
            "%{portforwards}": " ".join(["-portforward=" + p for p in ctx.attr.portforward_services]),
            "%{push_statements}": push_statements,
            "%{set_namespace}": ctx.executable._set_namespace.short_path,
            "%{it_manifest_filter}": ctx.executable._it_manifest_filter.short_path,
            "%{statements}": "\n".join(commands),
            "%{test_timeout}": ctx.attr.setup_timeout,
            "%{waitforapps}": " ".join(["-waitforapp=" + p for p in ctx.attr.wait_for_apps]),
        },
        output = ctx.outputs.executable,
    )

    rf = ctx.runfiles(files = files, transitive_files = depset(transitive = transitive))
    for dep_rf in pushes_runfiles:
        rf = rf.merge(dep_rf)
    return [DefaultInfo(
        executable = ctx.outputs.executable,
        runfiles = rf,
    )]

k8s_test_setup = rule(
    attrs = {
        "kubeconfig": attr.label(
            default = Label("@k8s_test//:kubeconfig"),
            allow_single_file = True,
        ),
        "kubectl": attr.label(
            default = Label("@k8s_test//:kubectl"),
            cfg = "host",
            executable = True,
            allow_single_file = True,
        ),
        "objects": attr.label_list(
            cfg = "target",
        ),
        "portforward_services": attr.string_list(),
        "setup_timeout": attr.string(default = "10m"),
        "wait_for_apps": attr.string_list(),
        "_it_sidecar": attr.label(
            default = Label("//testing/it_sidecar:it_sidecar"),
            cfg = "host",
            executable = True,
        ),
        "_kustomize": attr.label(
            default = Label("@kustomize_bin//:kustomize"),
            cfg = "host",
            executable = True,
        ),
        "_namespace_template": attr.label(
            default = Label("//skylib:k8s_test_namespace.sh.tpl"),
            allow_single_file = True,
        ),
        "_set_namespace": attr.label(
            default = Label("//skylib/kustomize:set_namespace"),
            cfg = "host",
            executable = True,
        ),
        "_it_manifest_filter": attr.label(
            default = Label("//it_manifest_filter:it_manifest_filter"),
            cfg = "host",
            executable = True,
        ),
        "_stamper": attr.label(
            default = Label("@io_bazel_rules_k8s//k8s:stamper"),
            cfg = "host",
            executable = True,
            allow_files = True,
        ),
        "_template_engine": attr.label(
            default = Label("//templating:fast_template_engine"),
            executable = True,
            cfg = "host",
        ),
    },
    executable = True,
    implementation = _k8s_test_setup_impl,
)
