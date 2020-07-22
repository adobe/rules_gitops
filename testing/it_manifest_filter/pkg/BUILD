load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["filter.go"],
    importpath = "github.com/adobe/rules_gitops/testing/it_manifest_filter/pkg",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/ghodss/yaml:go_default_library",
        "//vendor/k8s.io/api/apps/v1:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/yaml:go_default_library",
        "//vendor/k8s.io/client-go/util/jsonpath:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["filter_test.go"],
    data = glob(["testdata/**"]),
    embed = [":go_default_library"],
    deps = ["//vendor/github.com/google/go-cmp/cmp:go_default_library"],
)
