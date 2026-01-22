load("@io_bazel_rules_docker//container:providers.bzl", "PushInfo", "ImageInfo")
load("@rules_gitops//adapters:providers.bzl", "K8sPushInfo")

def _image_descriptor_impl(ctx):
    digestfile = ctx.attr.image[ImageInfo].container_parts["digest"]

    return [
        DefaultInfo(),
        K8sPushInfo(
            image_label = ctx.attr.image.label,
            registry = ctx.attr.registry,
            repository = ctx.attr.repository,
            digestfile = digestfile,
            pusher = ctx.attr.push[DefaultInfo],
        )
    ]


image_descriptor = rule(
    implementation = _image_descriptor_impl,
    attrs = {
        "image": attr.label(mandatory = True, providers = [ImageInfo]),
        "push": attr.label(mandatory = True, cfg = "target", providers = [PushInfo]),
        "registry": attr.string(mandatory = True),
        "repository": attr.string(mandatory = True),
    }
)