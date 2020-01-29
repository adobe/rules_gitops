load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "fields.go",
        "yaml.go",
        "yaml_go110.go",
    ],
    importmap = "git.corp.adobe.com/AdCloud/bazel_gitops_tools/vendor/github.com/ghodss/yaml",
    importpath = "github.com/ghodss/yaml",
    visibility = ["//visibility:public"],
    deps = ["//vendor/gopkg.in/yaml.v2:go_default_library"],
)
