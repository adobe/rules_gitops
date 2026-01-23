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
    files.append(namespace_create)

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
            cfg = "exec",
            executable = True,
            allow_single_file = True,
        ),
        "_namespace_template": attr.label(
            default = Label("//testing/private:k8s_test_namespace.sh.tpl"),
            allow_single_file = True,
        ),
    },
    executable = True,
    implementation = _k8s_test_namespace_impl,
)
