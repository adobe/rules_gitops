load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["it_manifest_filter.go"],
    importpath = "github.com/adobe/rules_gitops/testing/it_manifest_filter",
    visibility = ["//visibility:private"],
    deps = ["//testing/it_manifest_filter/pkg:go_default_library"],
)

go_binary(
    name = "it_manifest_filter",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
