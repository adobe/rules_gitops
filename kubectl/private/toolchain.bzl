"Kubectl toolchain rule"

load("//kubectl/private:providers.bzl", "KubectlToolchainInfo")

# Avoid using non-normalized paths (workspace/../other_workspace/path)
def _to_manifest_path(ctx, file):
    if file.short_path.startswith("../"):
        return "external/" + file.short_path[3:]
    else:
        return ctx.workspace_name + "/" + file.short_path

def _kubectl_toolchain_impl(ctx):
    executable = ctx.file.executable
    target_tool_path = _to_manifest_path(ctx, executable)

    template_variables = platform_common.TemplateVariableInfo({
        "KUBECTL_BIN": target_tool_path,
    })

    default = DefaultInfo(
        files = depset([executable]),
        runfiles = ctx.runfiles(files = [executable]),
    )
    kubectlinfo = KubectlToolchainInfo(
        target_tool_path = target_tool_path,
        executable = executable,
    )
    toolchain_info = platform_common.ToolchainInfo(
        kubectlinfo = kubectlinfo,
        template_variables = template_variables,
        default = default,
    )
    return [
        default,
        toolchain_info,
        template_variables,
    ]

kubectl_toolchain = rule(
    implementation = _kubectl_toolchain_impl,
    attrs = {
        "executable": attr.label(
            doc = "A hermetically downloaded executable target for the target platform.",
            mandatory = True,
            allow_single_file = True,
        ),
    },
)
