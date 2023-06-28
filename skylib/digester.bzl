load(
    "@io_bazel_rules_docker//container:layer_tools.bzl",
    _gen_img_args = "generate_args_for_image",
    _get_layers = "get_from_target",
)

def calc_digest(ctx):
    """ Calculate the digest for a given image

    Args:
       ctx: The context
    Returns:
        image: The image which has reference to a digest
    """

    digester_args = []
    digester_input = []

    image = _get_layers(ctx, ctx.label.name, ctx.attr.image)
    digester_img_args, digester_img_inputs = _gen_img_args(ctx, image)
    digester_input += digester_img_inputs
    digester_args += digester_img_args

    digester_args += ["--dst", str(ctx.outputs.digest.path), "--format", str(ctx.attr.format)]

    ctx.actions.run(
        inputs = digester_input,
        outputs = [ctx.outputs.digest],
        executable = ctx.executable._digester,
        arguments = digester_args,
        tools = ctx.attr._digester[DefaultInfo].default_runfiles.files,
        mnemonic = "CalculateDigest",
    )

    return image

def _calculate_digest_impl(ctx):
    calc_digest(ctx)

# Rule definition
calculate_digest = rule(
    doc = """
    Calculates the digest for a container image.

    Args:
      name: name of the rule
      image: the label of the image to calculate the digest for.
      format: The format to process: Docker or OCI.
    """,
    implementation = _image_digest_impl,
    attrs = {
        "format": attr.string(
            default = "Docker",
            values = [
                "OCI",
                "Docker",
            ],
            doc = "The format to process: Docker or OCI, default to 'Docker'.",
        ),
        "image": attr.label(
            mandatory = True,
            doc = "The label of the image to calculate the digest for.",
        ),
        "_digester": attr.label(
            default = "@io_bazel_rules_docker//container/go/cmd/digester",
            cfg = "exec",
            executable = True,
        ),
    },
    outputs = {
        "digest": "%{name}.digest",
    },
)
