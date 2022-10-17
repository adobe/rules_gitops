load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "generated.pb.go",
        "instr_fuzz.go",
        "intstr.go",
    ],
    importmap = "github.com/adobe/rules_gitops/vendor/k8s.io/apimachinery/pkg/util/intstr",
    importpath = "k8s.io/apimachinery/pkg/util/intstr",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/google/gofuzz:go_default_library",
        "//vendor/k8s.io/klog/v2:go_default_library",
        "@com_github_gogo_protobuf//proto:go_default_library",
    ],
)
