KustomizeToolchainInfo = provider(
    doc = "Toolchain information about the kustomize executable",
    fields = {
        "target_tool_path": "Path to the kustomize executable for the target platform.",
        "executable": "Hermetically download toolchain executable file",
    },
)

KustomizeInfo = provider(
    fields = {
        "image_pushes": "depset of image push executables for images referenced by this kustomization",
    },
)
