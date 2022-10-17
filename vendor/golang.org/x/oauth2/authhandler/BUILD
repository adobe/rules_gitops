load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["authhandler.go"],
    importmap = "github.com/adobe/rules_gitops/vendor/golang.org/x/oauth2/authhandler",
    importpath = "golang.org/x/oauth2/authhandler",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/oauth2:go_default_library"],
)
