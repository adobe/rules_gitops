load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["analysis.pb.go"],
    importpath = "github.com/adobe/rules_gitops/gitops/analysis",
    visibility = ["//visibility:public"],
    deps = [
        "//gitops/blaze_query:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
    ],
)
