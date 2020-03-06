load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["resolver.go"],
    importpath = "github.com/adobe/rules_gitops/resolver",
    visibility = ["//visibility:private"],
    deps = ["//resolver/pkg:go_default_library"],
)

go_binary(
    name = "resolver",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
