load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "apic.go",
        "decode.go",
        "emitterc.go",
        "encode.go",
        "parserc.go",
        "readerc.go",
        "resolve.go",
        "scannerc.go",
        "sorter.go",
        "writerc.go",
        "yaml.go",
        "yamlh.go",
        "yamlprivateh.go",
    ],
    importmap = "git.corp.adobe.com/AdCloud/bazel_gitops_tools/vendor/gopkg.in/yaml.v2",
    importpath = "gopkg.in/yaml.v2",
    visibility = ["//visibility:public"],
)
