"Repository rule to selectively determine which preset file is active"

def _preset_impl(rctx):
    version = rctx.getenv("USE_BAZEL_VERSION")

    preset_bzl_content = ""

    if version and version.startswith("8"):
        preset_bzl_content += """PRESET_FILE = "preset8.bazelrc\""""
    elif version and version.startswith("9"):
        preset_bzl_content += """PRESET_FILE = "preset9.bazelrc\""""
    else:
        preset_bzl_content += """PRESET_FILE = "preset7.bazelrc\""""

    rctx.file("BUILD.bazel", """exports_files(["preset.bzl"])""")
    rctx.file("preset.bzl", preset_bzl_content)

preset = repository_rule(
    implementation = _preset_impl,
)
