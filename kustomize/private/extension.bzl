"""Module extension for kustomize toolchain."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("//kustomize/private:platforms.bzl", "PLATFORMS")
load("//kustomize/private/versions:versions.bzl", "LATEST_KUSTOMIZE_VERSION", "VERSIONS")

def _kustomize_hub_impl(rctx):
    kustomize_hub_build_content = """
package(
    default_visibility = ["//visibility:public"]
)
"""
    toolchains_build_content = ""

    for [platform, meta] in PLATFORMS.items():
        toolchains_build_content += """
toolchain(
    name = "{platform}_toolchain",
    exec_compatible_with = {compatible_with},
    toolchain = "@kustomize_{platform}//:kustomize_toolchain",
    toolchain_type = "@rules_gitops//kustomize:toolchain_type",
)
""".format(
            platform = platform,
            compatible_with = meta.compatible_with,
        )

    rctx.file("toolchains/BUILD.bazel", content = toolchains_build_content)
    rctx.file("BUILD.bazel", content = kustomize_hub_build_content)

_kustomize_hub = repository_rule(
    implementation = _kustomize_hub_impl,
    attrs = {},
)

KUSTOMIZE_TOOLCHAIN_BUILD = """load("@rules_gitops//kustomize:defs.bzl", "kustomize_toolchain")
kustomize_toolchain(
    name = "kustomize_toolchain",
    executable = "kustomize"
)
"""

def _kustomize_extension_impl(module_ctx):
    kustomize_version = LATEST_KUSTOMIZE_VERSION

    for mod in module_ctx.modules:
        if len(mod.tags.toolchain) > 1:
            fail("Expected kustomize toolchain to only be declared once")

        # Allow the root module to override the kustomize toolchain version
        if mod.is_root and len(mod.tags.toolchain) == 1:
            kustomize_version = mod.tags.toolchain[0].version

    if not VERSIONS[kustomize_version]:
        fail("No matching version found for kustomize v{}".format(kustomize_version))

    binaries = VERSIONS[kustomize_version]

    for [platform, sha256] in binaries.items():
        http_archive(
            name = "kustomize_{}".format(platform),
            urls = [
                "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v{version}/kustomize_v{version}_{platform}.tar.gz".format(version = kustomize_version, platform = platform.replace("-", "_")),
            ],
            build_file_content = KUSTOMIZE_TOOLCHAIN_BUILD,
            sha256 = sha256,
        )

    _kustomize_hub(
        name = "kustomize",
    )

    return module_ctx.extension_metadata()

kustomize = module_extension(
    implementation = _kustomize_extension_impl,
    tag_classes = {
        "toolchain": tag_class(
            attrs = {
                "version": attr.string(
                    doc = "kustomize binary version",
                ),
            },
        ),
    },
)
