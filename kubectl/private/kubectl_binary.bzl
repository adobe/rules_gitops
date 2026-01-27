"""
Simple rule for running kubectl from the toolchain config
"""

load("//adapters:providers.bzl", "K8sPushInfo")
load("//kustomize:defs.bzl", "KustomizeInfo")
load("//stamper:stamp.bzl", "stamp")

def _kubectl_binary_impl(ctx):
    executable = ctx.toolchains["@rules_gitops//kubectl:toolchain_type"].kubectlinfo.executable

    runfiles = ctx.runfiles(
        files = ctx.files.srcs + [ctx.executable._template_engine, ctx.file._info_file, executable],
    ).merge(ctx.attr._bash_runfiles[DefaultInfo].default_runfiles)

    files = []

    cluster_arg = ctx.attr.cluster
    cluster_arg = ctx.expand_make_variables("cluster", cluster_arg, {})
    if "{" in ctx.attr.cluster:
        cluster_arg, cluster_files = stamp(ctx, cluster_arg, ctx.files.srcs, ctx.label.name + ".cluster-name")
        files.extend(cluster_files)

    user_arg = ctx.attr.user
    user_arg = ctx.expand_make_variables("user", user_arg, {})
    if "{" in ctx.attr.user:
        user_arg, user_files = stamp(ctx, user_arg, ctx.files.srcs, ctx.label.name + ".user-name")
        files.extend(user_files)

    kubectl_command_arg = ctx.attr.command
    kubectl_command_arg = ctx.expand_make_variables("kubectl_command", kubectl_command_arg, {})

    statements = ""
    transitive_runfiles = []

    if ctx.attr.push:
        trans_img_pushes = depset(transitive = [obj[KustomizeInfo].image_pushes for obj in ctx.attr.srcs]).to_list()
        statements += "\n".join([
            "# {}\n".format(exe[K8sPushInfo].image_label) +
            "echo  pushing {}/{}".format(exe[K8sPushInfo].registry, exe[K8sPushInfo].repository)
            for exe in trans_img_pushes
            if hasattr(exe[K8sPushInfo], "pusher")
        ]) + "\n"
        statements += "\n".join([
            "async \"%s\"" % exe[K8sPushInfo].pusher.files_to_run.executable.short_path
            for exe in trans_img_pushes
            if hasattr(exe[K8sPushInfo], "pusher")
        ]) + "\nwaitpids\n"

        transitive_runfiles += [exe[K8sPushInfo].pusher.default_runfiles for exe in trans_img_pushes if hasattr(exe[K8sPushInfo], "pusher")]

    runfiles = runfiles.merge_all(transitive_runfiles)

    namespace = ctx.attr.namespace
    for inattr in ctx.attr.srcs:
        for infile in inattr.files.to_list():
            statements += "{template_engine} --template={infile} --variable=NAMESPACE={namespace} --stamp_info_file={info_file} | \"{kubectl}\" --cluster=\"{cluster}\" --user=\"{user}\" {kubectl_command} -f -\n".format(
                kubectl = executable.short_path,
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

    files.append(ctx.outputs.executable)

    return [
        DefaultInfo(files = depset(files), runfiles = runfiles, executable = ctx.outputs.executable),
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
        "_bash_runfiles": attr.label(
            default = Label("@bazel_tools//tools/bash/runfiles"),
        ),
    },
    executable = True,
    implementation = _kubectl_binary_impl,
    toolchains = ["@rules_gitops//kubectl:toolchain_type"],
)
