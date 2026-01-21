load("//kustomize:defs.bzl", "KustomizeInfo")
load("//adapters:providers.bzl", "K8sPushInfo")

def _image_push_statements(
        ctx,
        kustomize_objs,
        files = []):
    statements = ""
    trans_img_pushes = depset(transitive = [obj[KustomizeInfo].image_pushes for obj in kustomize_objs]).to_list()
    statements += "\n".join([
        "echo  pushing {}/{}".format(exe[K8sPushInfo].registry, exe[K8sPushInfo].repository)
        for exe in trans_img_pushes
    ]) + "\n"
    statements += "\n".join([
        "async \"%s\"" % exe.files_to_run.executable.short_path
        for exe in trans_img_pushes
    ]) + "\nwaitpids\n"
    files += [obj.files_to_run.executable for obj in trans_img_pushes]
    dep_runfiles = [obj[DefaultInfo].default_runfiles for obj in trans_img_pushes]
    return statements, files, dep_runfiles

def _k8s_test_setup_impl(ctx):
    files = []  # runfiles list
    transitive = []
    commands = []  # the list of commands to execute

    kustomize_executable = ctx.toolchains["@rules_gitops//kustomize:toolchain_type"].kustomizeinfo.executable

    # add files referenced by rule attributes to runfiles
    files = [ctx.executable._stamper, ctx.file.kubectl, ctx.file.kubeconfig, kustomize_executable, ctx.executable._it_sidecar, ctx.executable._it_manifest_filter]
    files += ctx.files._set_namespace
    files += ctx.files.cluster

    push_statements, files, pushes_runfiles = _image_push_statements(ctx, [o for o in ctx.attr.objects if KustomizeInfo in o], files)

    # execute all objects targets
    for obj in ctx.attr.objects:
        if obj.files_to_run.executable:
            # add object' targets and excutables to runfiles
            files.append(obj.files_to_run.executable)
            transitive.append(obj.default_runfiles.files)

            # add object' execution command
            commands.append(obj.files_to_run.executable.short_path + " | ${SET_NAMESPACE} $NAMESPACE | ${IT_MANIFEST_FILTER} | ${KUBECTL} apply -f -")
        else:
            files += obj.files.to_list()
            commands += [ctx.executable._template_engine.short_path + " --template=" + filename.short_path + " --variable=NAMESPACE=${NAMESPACE} | ${SET_NAMESPACE} $NAMESPACE | ${IT_MANIFEST_FILTER} | ${KUBECTL} apply -f -" for filename in obj.files.to_list()]

    files.append(ctx.executable._template_engine)

    # create namespace script
    ctx.actions.expand_template(
        template = ctx.file._namespace_template,
        substitutions = {
            "%{it_sidecar}": ctx.executable._it_sidecar.short_path,
            "%{cluster}": ctx.file.cluster.path,
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
    rf = rf.merge(ctx.attr._set_namespace[DefaultInfo].default_runfiles)
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
            cfg = "exec",
            executable = True,
            allow_single_file = True,
        ),
        "objects": attr.label_list(
            cfg = "target",
        ),
        "portforward_services": attr.string_list(),
        "setup_timeout": attr.string(default = "10m"),
        "wait_for_apps": attr.string_list(),
        "cluster": attr.label(
            default = Label("@k8s_test//:cluster"),
            allow_single_file = True,
        ),
        "_it_sidecar": attr.label(
            default = Label("//testing/it_sidecar:it_sidecar"),
            cfg = "exec",
            executable = True,
        ),
        "_namespace_template": attr.label(
            default = Label("//testing/private:k8s_test_namespace.sh.tpl"),
            allow_single_file = True,
        ),
        "_set_namespace": attr.label(
            default = Label("//testing/private:set_namespace"),
            cfg = "exec",
            executable = True,
        ),
        "_it_manifest_filter": attr.label(
            default = Label("//testing/it_manifest_filter:it_manifest_filter"),
            cfg = "exec",
            executable = True,
        ),
        "_stamper": attr.label(
            default = Label("//stamper:stamper"),
            cfg = "exec",
            executable = True,
            allow_files = True,
        ),
        "_template_engine": attr.label(
            default = Label("//templating:fast_template_engine"),
            executable = True,
            cfg = "exec",
        ),
    },
    executable = True,
    implementation = _k8s_test_setup_impl,
    toolchains = ["@rules_gitops//kustomize:toolchain_type"]
)
