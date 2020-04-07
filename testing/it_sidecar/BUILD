load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["it_sidecar.go"],
    importpath = "github.com/adobe/rules_gitops/testing/it_sidecar",
    visibility = ["//visibility:private"],
    deps = [
        "//testing/it_sidecar/stern:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/client-go/informers:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/rest:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/k8s.io/client-go/tools/clientcmd:go_default_library",
        "//vendor/k8s.io/client-go/tools/portforward:go_default_library",
        "//vendor/k8s.io/client-go/transport/spdy:go_default_library",
        "//vendor/k8s.io/client-go/util/homedir:go_default_library",
    ],
)

go_binary(
    name = "it_sidecar",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
