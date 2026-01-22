K8sPushInfo = provider(
    doc = "Information required to inject image into a manifest and optionally push it",
    fields = {
        "image_label":  "bazel target label of the image",
        "legacy_image_name": "optional short name",
        "registry": "registry where the image will live",
        "repository": "repository where the image will live",
        "digestfile": "a file containing the digest of the image",
        "pusher": "an executable target used to push the image to a remote registry",
        "run_environment": "a run environment info provider used for pushing the image",
    },
)