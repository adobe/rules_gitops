"""
Simple rule for running kubectl from the toolchain config
"""
load("//kustomize:defs.bzl", "KustomizeInfo")
load("//adapters:providers.bzl", "K8sPushInfo")
load("//stamper:stamp.bzl", "stamp")

def _kubectl_binary_impl(ctx):
    files = [] + ctx.files.srcs

    executable = ctx.toolchains["@rules_gitops//kubectl:toolchain_type"].kubectlinfo.executable

    cluster_arg = ctx.attr.cluster
    cluster_arg = ctx.expand_make_variables("cluster", cluster_arg, {})
    if "{" in ctx.attr.cluster:
        cluster_arg = stamp(ctx, cluster_arg, files, ctx.label.name + ".cluster-name", True)

    user_arg = ctx.attr.user
    user_arg = ctx.expand_make_variables("user", user_arg, {})
    if "{" in ctx.attr.user:
        user_arg = stamp(ctx, user_arg, files, ctx.label.name + ".user-name", True)

    kubectl_command_arg = ctx.attr.command
    kubectl_command_arg = ctx.expand_make_variables("kubectl_command", kubectl_command_arg, {})

    statements = ""
    transitive = None
    transitive_runfiles = []

    files += [ctx.executable._template_engine, ctx.file._info_file]

    if ctx.attr.push:
        trans_img_pushes = depset(transitive = [obj[KustomizeInfo].image_pushes for obj in ctx.attr.srcs]).to_list()
        statements += "\n".join([
            "# {}\n".format(exe[K8sPushInfo].image_label) +
            "echo  pushing {}/{}".format(exe[K8sPushInfo].registry, exe[K8sPushInfo].repository)
            for exe in trans_img_pushes
        ]) + "\n"
        statements += "\n".join([
            "async \"%s\"" % exe.files_to_run.executable.short_path
            for exe in trans_img_pushes
        ]) + "\nwaitpids\n"
        files += [obj.files_to_run.executable for obj in trans_img_pushes]
        transitive = depset(transitive = [obj.default_runfiles.files for obj in trans_img_pushes])
        transitive_runfiles += [exe[DefaultInfo].default_runfiles for exe in trans_img_pushes]

    namespace = ctx.attr.namespace
    for inattr in ctx.attr.srcs:
        for infile in inattr.files.to_list():
            statements += "{template_engine} --template={infile} --variable=NAMESPACE={namespace} --stamp_info_file={info_file} | kubectl --cluster=\"{cluster}\" --user=\"{user}\" {kubectl_command} -f -\n".format(
                infile = infile.short_path,
                cluster = cluster_arg,
                user = user_arg,
                kubectl_command = kubectl_command_arg,
                template_engine = ctx.executable._template_engine.short_path,
                namespace = namespace,
                info_file = ctx.file._info_file.short_path,
            )

    ctx.actions.expand_template(
        template = ctx.file._template,
        substitutions = {
            "%{statements}": statements,
        },
        output = ctx.outputs.executable,
    )

    runfiles = ctx.runfiles(files = files, transitive_files = transitive)
    runfiles = runfiles.merge_all(transitive_runfiles)

    return [
        DefaultInfo(runfiles = runfiles),
    ]

kubectl_binary = rule(
    attrs = {
        "srcs": attr.label_list(providers = [KustomizeInfo]),
        "cluster": attr.string(mandatory = True),
        "namespace": attr.string(mandatory = True),
        "command": attr.string(default = "apply"),
        "user": attr.string(default = "{BUILD_USER}"),
        "push": attr.bool(default = True),
        "_build_user_value": attr.label(
            default = Label("//stamper:build_user_value.txt"),
            allow_single_file = True,
        ),
        "_info_file": attr.label(
            default = Label("//stamper:more_stable_status.txt"),
            allow_single_file = True,
        ),
        "_stamper": attr.label(
            default = Label("//stamper:stamper"),
            cfg = "exec",
            executable = True,
            allow_files = True,
        ),
        "_template": attr.label(
            default = Label("//kubectl/private:run-all.sh.tpl"),
            allow_single_file = True,
        ),
        "_template_engine": attr.label(
            default = Label("//templating:fast_template_engine"),
            executable = True,
            cfg = "exec",
        ),
    },
    executable = True,
    implementation = _kubectl_binary_impl,
    toolchains = ["@rules_gitops//kubectl:toolchain_type"],
)