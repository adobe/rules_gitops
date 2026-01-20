"""
Simple rule for running kustomize from the toolchain config
"""

def _kustomize_binary(ctx):
    executable = ctx.toolchains["@rules_gitops//kustomize:toolchain_type"].kustomizeinfo.executable

    kustomize = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.symlink(
        output = kustomize,
        target_file = executable,
        is_executable = True,
    )

    return [
        DefaultInfo(
            runfiles = ctx.runfiles(files = [executable, kustomize]),
            executable = kustomize,
        ),
    ]

kustomize_binary = rule(
    implementation = _kustomize_binary,
    attrs = {},
    toolchains = ["@rules_gitops//kustomize:toolchain_type"],
    executable = True,
)
