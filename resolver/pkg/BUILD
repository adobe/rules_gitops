load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["resolver.go"],
    importpath = "github.com/adobe/rules_gitops/resolver/pkg",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_ghodss_yaml//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1/unstructured:go_default_library",
        "@io_k8s_apimachinery//pkg/util/yaml:go_default_library",
    ],
)

# TODO(KZ): this test uses private manifests
go_test(
    name = "go_default_test",
    srcs = ["resolver_test.go"],
    data = glob(["testdata/**"]),
    embed = [":go_default_library"],
)
