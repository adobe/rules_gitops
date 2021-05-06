load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["wire.go"],
    importmap = "github.com/adobe/rules_gitops/vendor/google.golang.org/protobuf/encoding/protowire",
    importpath = "google.golang.org/protobuf/encoding/protowire",
    visibility = ["//visibility:public"],
    deps = ["//vendor/google.golang.org/protobuf/internal/errors:go_default_library"],
)
