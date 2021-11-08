load("@io_bazel_rules_go//go:def.bzl", "go_test")

go_test(
    name = "go_default_test",
    srcs = ["sidecar_client_test_no_callback.go"],
    args = [
        "-setup",
        "$(location //testing/it_sidecar/client:noop.setup)",
    ],
    data = [
        "//testing/it_sidecar/client:noop.setup",
    ],
    rundir = ".",
    deps = ["//testing/it_sidecar/client:go_default_library"],
)
