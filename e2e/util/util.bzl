load("@rules_shell//shell:sh_test.bzl", "sh_test")
load("@rules_shell//shell:sh_binary.bzl", "sh_binary")
load("@bazel_lib//lib:expand_template.bzl", "expand_template")

def e2e_test(name, steps):
    """Declares an e2e test interacting with a kind cluster
    
    Args:
        name: name of the test target
        setups: a list of labels producing binaries to run
    """

    sh_test(
        name = "manual_k8s",
        srcs = ["//e2e/util:test.sh"],
        args = ["$(rlocationpath {label})".format(label = label) for label in steps],
        data = [
            "//e2e/util:setup",
            "@io_k8s_sigs_kind//:kind",
            "//kubectl:resolved_toolchain",
            "//e2e/util:teardown",
        ] + steps,
        env = {
            "SETUP": "$(rlocationpath //e2e/util:setup)",
            "TEARDOWN": "$(rlocationpath //e2e/util:teardown)",
            "KIND_BIN_PATH": "$(rlocationpath @io_k8s_sigs_kind//:kind)",
            "KUBECTL_BIN_PATH": "$(rlocationpath //kubectl:resolved_toolchain)",
        },
        deps = [
            "@bazel_tools//tools/bash/runfiles",
        ],
    )


def kubectl_cmd(name, args):
    expand_template(
        name = name + ".script",
        is_executable = True,
        template = "//e2e/util:kubectl_cmd.sh.tpl",
        substitutions = {
            "$$KUBECTL_ARGS$$": " ".join(args)
        },
        out = name + ".bash"
    )

    sh_binary(
        name = name,
        srcs = [name + ".script"],
        env = {
            "KUBECTL_BIN_PATH": "$(rlocationpath //kubectl:resolved_toolchain)"
        },
        data = [
            "//kubectl:resolved_toolchain"
        ],
        deps = [
            "@bazel_tools//tools/bash/runfiles",
        ],
        toolchains = [
            "//kubectl:resolved_toolchain"
        ],
    )