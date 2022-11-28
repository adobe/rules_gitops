"""
Implementation of external image information provider suitable for injection into manifests
"""

load(":push.bzl", "K8sPushInfo")

def _external_image_impl(ctx):
    sv = ctx.attr.image.split("@", 1)
    if (len(sv) == 1) and (not ctx.attr.digest):
        fail("digest must be specified either in image or as a separate attribute")
    s = sv[0].split(":", 1)[0]  #drop tag
    registry, repository = s.split("/", 1)

    #write digest to a file
    digest_file = ctx.actions.declare_file(ctx.label.name + ".digest")
    ctx.actions.write(
        output = digest_file,
        content = ctx.attr.digest,
    )
    return [
        DefaultInfo(
            files = depset([digest_file]),
        ),
        K8sPushInfo(
            image_label = ctx.label,
            legacy_image_name = ctx.attr.image_name,
            registry = registry,
            repository = repository,
            digestfile = digest_file,
        ),
    ]

external_image = rule(
    implementation = _external_image_impl,
    attrs = {
        "image": attr.string(mandatory = True, doc = "The image location, e.g. gcr.io/foo/bar:baz"),
        "image_name": attr.string(doc = "Image name, e.g. exernalserver. DEPRECATED: Use full target label instead, e.g. //images:externalserver"),
        "digest": attr.string(mandatory = True, doc = "The image digest, e.g. sha256:deadbeef"),
    },
)
