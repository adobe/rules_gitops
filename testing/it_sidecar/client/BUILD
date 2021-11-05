load("@io_bazel_rules_go//go:def.bzl", "go_library")

exports_files(["noop.setup"])

go_library(
    name = "go_default_library",
    srcs = ["sidecar_client.go"],
    data = [
        ":noop.setup",
    ],
    importpath = "github.com/adobe/rules_gitops/testing/it_sidecar/client",
    visibility = ["//visibility:public"],
)
