load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "fuzz.go",
    ],
    importmap = "git.corp.adobe.com/AdCloud/bazel_gitops_tools/vendor/github.com/google/gofuzz",
    importpath = "github.com/google/gofuzz",
    visibility = ["//visibility:public"],
)
