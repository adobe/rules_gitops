load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["sidecar_client.go"],
    importpath = "github.com/adobe/rules_gitops/testing/it_sidecar/client",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["sidecar_client_test.go"],
    args = [
        "-setup",
        "$(location :noop.setup)",
    ],
    data = [
        ":noop.setup",
    ],
    embed = [":go_default_library"],
    rundir = ".",
)
