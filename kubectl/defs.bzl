load("//kubectl/private:extension.bzl", _kubectl = "kubectl")
load("//kubectl/private:toolchain.bzl", _kubectl_toolchain = "kubectl_toolchain")
load("//kubectl/private:kubectl_binary.bzl", _kubectl_binary = "kubectl_binary")

kubectl = _kubectl
kubectl_binary = _kubectl_binary
kubectl_toolchain = _kubectl_toolchain
