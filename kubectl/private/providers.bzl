"""Provider definitions for kubectl toolchain."""

KubectlToolchainInfo = provider(
    doc = "Toolchain information about the kubectl executable",
    fields = {
        "target_tool_path": "Path to the kubectl executable for the target platform.",
        "executable": "Hermetically download toolchain executable file",
    },
)
