"""Adapter for rules_img container images."""

load("@rules_img//img:providers.bzl", "DeployInfo", "ImageIndexInfo", "ImageManifestInfo")
load("//adapters:providers.bzl", "K8sPushInfo")

def _k8s_push_info_impl(ctx):
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

k8s_push_info = rule(
    implementation = _k8s_push_info_impl,
    attrs = {
        "image": attr.label(mandatory = True, providers = [[ImageManifestInfo], [ImageIndexInfo]]),
        "push": attr.label(mandatory = True, cfg = "target", providers = [DeployInfo]),
        "registry": attr.string(mandatory = True),
        "repository": attr.string(mandatory = True),
    },
)
