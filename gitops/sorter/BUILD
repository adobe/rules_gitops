load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/adobe/rules_gitops/gitops/sorter",
    visibility = ["//visibility:private"],
    deps = ["//gitops/k8ssort:go_default_library"],
)

go_binary(
    name = "sorter",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
