# Bazel GitOps Rules Example Repository

## Overview

The example Bazel project to demonstrate Bazel GitOps Rules use:

- [WORKSPACE](./WORKSPACE) -- the minimal workspace setup file
- [helloworld](./helloworld) -- the minimal Go application with `k8s_deploy` manifests

All following commands assume that `/examples` is the current directory.

## Build & Test

```
bazel test //...
```

## Render Helloworld Application Deployment Manifests

```
bazel run //helloworld:mynamespace.show
```

