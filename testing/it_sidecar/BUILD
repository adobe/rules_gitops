load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["it_sidecar.go"],
    importpath = "github.com/adobe/rules_gitops/testing/it_sidecar",
    visibility = ["//visibility:private"],
    deps = [
        "//testing/it_sidecar/stern:go_default_library",
        "@io_k8s_api//core/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_client_go//informers:go_default_library",
        "@io_k8s_client_go//kubernetes:go_default_library",
        "@io_k8s_client_go//rest:go_default_library",
        "@io_k8s_client_go//tools/cache:go_default_library",
        "@io_k8s_client_go//tools/clientcmd:go_default_library",
        "@io_k8s_client_go//tools/portforward:go_default_library",
        "@io_k8s_client_go//transport/spdy:go_default_library",
        "@io_k8s_client_go//util/homedir:go_default_library",
    ],
)

go_binary(
    name = "it_sidecar",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
