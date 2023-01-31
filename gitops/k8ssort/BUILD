load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["sorter.go"],
    importpath = "github.com/adobe/rules_gitops/gitops/k8ssort",
    visibility = ["//visibility:public"],
    deps = ["//vendor/gopkg.in/yaml.v2:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["sorter_test.go"],
    embed = [":go_default_library"],
    deps = ["//vendor/github.com/google/go-cmp/cmp:go_default_library"],
)
