load("@rules_img//img:providers.bzl", "DeployInfo", "ImageManifestInfo")
load("//adapters:providers.bzl", "K8sPushInfo")

def _image_descriptor_impl(ctx):
    digestfile = ctx.attr.image[OutputGroupInfo].digest.to_list()[0]

    return [
        DefaultInfo(),
        K8sPushInfo(
            image_label = ctx.attr.image.label,
            registry = ctx.attr.registry,
            repository = ctx.attr.repository,
            digestfile = digestfile,
            pusher = ctx.attr.push[DefaultInfo],
            run_environment = ctx.attr.push[RunEnvironmentInfo],
        ),
    ]

image_descriptor = rule(
    implementation = _image_descriptor_impl,
    attrs = {
        "image": attr.label(mandatory = True, providers = [ImageManifestInfo]),
        "push": attr.label(mandatory = True, cfg = "target", providers = [DeployInfo]),
        "registry": attr.string(mandatory = True),
        "repository": attr.string(mandatory = True),
    },
)
