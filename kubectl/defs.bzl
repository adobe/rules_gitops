load("//kubectl/private:extension.bzl", _kubectl = "kubectl")
load("//kubectl/private:kubectl_binary.bzl", _kubectl_binary = "kubectl_binary")
load("//kubectl/private:resolved_toolchain.bzl", _resolved_toolchain = "resolved_toolchain")
load("//kubectl/private:toolchain.bzl", _kubectl_toolchain = "kubectl_toolchain")

kubectl = _kubectl
kubectl_binary = _kubectl_binary
kubectl_toolchain = _kubectl_toolchain
resolved_toolchain = _resolved_toolchain
