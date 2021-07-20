load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "discard.go",
        "logr.go",
    ],
    importmap = "github.com/adobe/rules_gitops/vendor/github.com/go-logr/logr",
    importpath = "github.com/go-logr/logr",
    visibility = ["//visibility:public"],
)
