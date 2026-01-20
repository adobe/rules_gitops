"""
Simple rule for running kubectl from the toolchain config
"""

def _kubectl_binary(ctx):
    executable = ctx.toolchains["@rules_gitops//kubectl:toolchain_type"].kubectlinfo.executable

    kubectl = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.symlink(
        output = kubectl,
        target_file = executable,
        is_executable = True,
    )

    return [
        DefaultInfo(
            runfiles = ctx.runfiles(files = [executable, kubectl]),
            executable = kubectl,
        ),
    ]

kubectl_binary = rule(
    implementation = _kubectl_binary,
    attrs = {},
    toolchains = ["@rules_gitops//kubectl:toolchain_type"],
    executable = True,
)
