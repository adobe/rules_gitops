load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["gitlab.go"],
    importpath = "github.com/adobe/rules_gitops/gitops/git/gitlab",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/xanzy/go-gitlab:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["gitlab_test.go"],
    embed = [":go_default_library"],
)
