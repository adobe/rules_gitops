"""Public API for Kubernetes testing rules."""

load("//testing/private:k8s_test_namespace.bzl", _k8s_test_namespace = "k8s_test_namespace")
load("//testing/private:k8s_test_setup.bzl", _k8s_test_setup = "k8s_test_setup")

k8s_test_namespace = _k8s_test_namespace
k8s_test_setup = _k8s_test_setup
