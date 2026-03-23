"""Module extension for kubectl toolchain."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")
load("//kubectl/private:platforms.bzl", "PLATFORMS")
load("//kubectl/private/versions:versions.bzl", "LATEST_KUBECTL_VERSION", "VERSIONS")

def _kubectl_hub_impl(rctx):
    kubectl_hub_build_content = """
package(
    default_visibility = ["//visibility:public"]
)
"""
    toolchains_build_content = """load("@rules_gitops//kubectl:defs.bzl", "kubectl_toolchain")"""

    for [platform, meta] in PLATFORMS.items():
        toolchains_build_content += """
kubectl_toolchain(
    name = "kubectl_{platform}_toolchain",
    executable = "@kubectl_{platform}//file"
)

toolchain(
    name = "{platform}_toolchain",
    exec_compatible_with = {compatible_with},
    toolchain = ":kubectl_{platform}_toolchain",
    toolchain_type = "@rules_gitops//kubectl:toolchain_type",
)
""".format(
            platform = platform,
            compatible_with = meta.compatible_with,
        )

    rctx.file("toolchains/BUILD.bazel", content = toolchains_build_content)
    rctx.file("BUILD.bazel", content = kubectl_hub_build_content)

_kubectl_hub = repository_rule(
    implementation = _kubectl_hub_impl,
    attrs = {},
)

KUBECTL_TOOLCHAIN_BUILD = """
kubectl_toolchain(
    name = "kubectl_toolchain",
    executable = "file"
)
"""

def _kubectl_extension_impl(module_ctx):
    kubectl_version = LATEST_KUBECTL_VERSION

    for mod in module_ctx.modules:
        if len(mod.tags.toolchain) > 1:
            fail("Expected kubectl toolchain to only be declared once")

        # Allow the root module to override the kubectl toolchain version
        if mod.is_root and len(mod.tags.toolchain) == 1:
            kubectl_version = mod.tags.toolchain[0].version

    if not VERSIONS[kubectl_version]:
        fail("No matching version found for kubectl v{}".format(kubectl_version))

    binaries = VERSIONS[kubectl_version]

    for [platform, sha256] in binaries.items():
        [os, arch] = platform.split("_")
        http_file(
            name = "kubectl_{}".format(platform),
            urls = [
                "https://dl.k8s.io/release/v{version}/bin/{os}/{arch}/kubectl".format(version = kubectl_version, os = os, arch = arch),
            ],
            executable = True,
            sha256 = sha256,
        )

    _kubectl_hub(
        name = "kubectl",
    )

    return module_ctx.extension_metadata()

kubectl = module_extension(
    implementation = _kubectl_extension_impl,
    tag_classes = {
        "toolchain": tag_class(
            attrs = {
                "version": attr.string(
                    doc = "kubectl binary version",
                ),
            },
        ),
    },
)
