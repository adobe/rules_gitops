load("//kustomize/private:extension.bzl", _kustomize = "kustomize")
load("//kustomize/private:toolchain.bzl", _kustomize_toolchain = "kustomize_toolchain")
load("//kustomize/private:kustomize_binary.bzl", _kustomize_binary = "kustomize_binary")
load("//kustomize/private:kustomization.bzl", _kustomization = "kustomization")

kustomize = _kustomize
kustomization = _kustomization
kustomize_binary = _kustomize_binary
kustomize_toolchain = _kustomize_toolchain
