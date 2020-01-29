load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "dec.go",
        "rounder.go",
    ],
    importmap = "git.corp.adobe.com/AdCloud/bazel_gitops_tools/vendor/gopkg.in/inf.v0",
    importpath = "gopkg.in/inf.v0",
    visibility = ["//visibility:public"],
)
