# Copyright 2026 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

"""Rule for building kustomization overlays."""

load("//adapters:providers.bzl", "K8sPushInfo")
load("//kustomize/private:providers.bzl", "KustomizeInfo")
load("//stamper:stamp.bzl", "stamp")

def _stamp_file(ctx, infile, output):
    stamps = [ctx.file._info_file]
    stamp_args = [
        "--stamp-info-file=%s" % sf.path
        for sf in stamps
    ]
    ctx.actions.run(
        executable = ctx.executable._stamper,
        arguments = [
            "--format-file=%s" % infile.path,
            "--output=%s" % output.path,
        ] + stamp_args,
        inputs = [infile] + stamps,
        outputs = [output],
        mnemonic = "Stamp",
        tools = [ctx.executable._stamper],
    )

def _is_ignored_src(src):
    basename = src.rsplit("/", 1)[-1]
    return basename.startswith(".")

_script_template = """\
#!/usr/bin/env bash
set -euo pipefail
{kustomize} build --load-restrictor LoadRestrictionsNone --reorder legacy {kustomize_dir} {template_part} {resolver_part} >{out}
"""

def _kustomization_impl(ctx):
    kustomization_yaml_file = ctx.actions.declare_file(ctx.attr.name + "/kustomization.yaml")
    root = kustomization_yaml_file.dirname

    upupup = "/".join([".."] * (root.count("/") + 1))
    use_stamp = False
    tmpfiles = []
    kustomization_yaml = "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\n"
    kustomization_yaml += "resources:\n"
    for _, f in enumerate(ctx.files.manifests):
        kustomization_yaml += "- {}/{}\n".format(upupup, f.path)

    if ctx.attr.namespace:
        kustomization_yaml += "namespace: '{}'\n".format(ctx.attr.namespace)
        use_stamp = use_stamp or "{" in ctx.attr.namespace

    if ctx.attr.name_prefix:
        kustomization_yaml += "namePrefix: '{}'\n".format(ctx.attr.name_prefix)
        use_stamp = use_stamp or "{" in ctx.attr.name_prefix

    if ctx.attr.name_suffix:
        kustomization_yaml += "nameSuffix: '{}'\n".format(ctx.attr.name_suffix)
        use_stamp = use_stamp or "{" in ctx.attr.name_suffix

    if ctx.attr.configurations:
        kustomization_yaml += "configurations:\n"
        for _, f in enumerate(ctx.files.configurations):
            kustomization_yaml += "- {}/{}\n".format(upupup, f.path)

    if ctx.files.patches:
        kustomization_yaml += "patches:\n"
        for _, f in enumerate(ctx.files.patches):
            # TODO: this changed in later versions of kustomize
            kustomization_yaml += "- path: {}/{}\n".format(upupup, f.path)

    if ctx.attr.image_name_patches or ctx.attr.image_tag_patches:
        kustomization_yaml += "images:\n"
        for image, new_tag in ctx.attr.image_tag_patches.items():
            new_name = ctx.attr.image_name_patches.get(image, default = None)
            kustomization_yaml += "- name: \"{}\"\n".format(image)
            kustomization_yaml += "  newTag: \"{}\"\n".format(new_tag)
            if new_name != None:
                kustomization_yaml += "  newName: \"{}\"\n".format(new_name)
        for image, new_name in ctx.attr.image_name_patches.items():
            if ctx.attr.image_tag_patches.get(image, default = None) == None:
                kustomization_yaml += "- name: \"{}\"\n".format(image)
                kustomization_yaml += "  newName: \"{}\"\n".format(new_name)

    if ctx.attr.common_labels:
        kustomization_yaml += "commonLabels:\n"
        for k in ctx.attr.common_labels:
            use_stamp = use_stamp or "{" in ctx.attr.common_labels[k]
            kustomization_yaml += "  {}: '{}'\n".format(k, ctx.attr.common_labels[k])

    if ctx.attr.common_annotations:
        kustomization_yaml += "commonAnnotations:\n"
        for k in ctx.attr.common_annotations:
            use_stamp = use_stamp or "{" in ctx.attr.common_annotations[k]
            kustomization_yaml += "  {}: '{}'\n".format(k, ctx.attr.common_annotations[k])

    kustomization_yaml += "generatorOptions:\n"

    kustomization_yaml += "  disableNameSuffixHash: {}\n".format(str(ctx.attr.disable_name_suffix_hash).lower())

    if ctx.attr.configmaps_srcs:
        maps = dict()  # configmap name to list of File objects
        for target in ctx.attr.configmaps_srcs:
            for src in target.files.to_list():
                # ignore dot files
                if _is_ignored_src(src.path):
                    continue
                mapname = src.path.rsplit("/")[-2]
                if not mapname in maps:
                    maps[mapname] = []
                maps[mapname].append(src)
        kustomization_yaml += "configMapGenerator:\n"
        for cmname in maps:
            kustomization_yaml += "- name: {}\n".format(cmname)
            kustomization_yaml += "  files:\n"
            for f in maps[cmname]:
                kustomization_yaml += "  - {}/{}\n".format(upupup, f.path)

    if ctx.attr.secrets_srcs:
        maps = dict()  # secret name to list of File objects
        for target in ctx.attr.secrets_srcs:
            for src in target.files.to_list():
                # ignore dot files
                if _is_ignored_src(src.path):
                    continue
                mapname = src.path.rsplit("/")[-2]
                if not mapname in maps:
                    maps[mapname] = []
                maps[mapname].append(src)
        kustomization_yaml += "secretGenerator:\n"
        for cmname in maps:
            kustomization_yaml += "- name: {}\n".format(cmname)
            kustomization_yaml += "  type: Opaque\n"
            kustomization_yaml += "  files:\n"
            for f in maps[cmname]:
                kustomization_yaml += "  - {}/{}\n".format(upupup, f.path)

    if use_stamp:
        kustomization_yaml_unstamped_file = ctx.actions.declare_file(ctx.attr.name + "/unstamped.yaml")
        ctx.actions.write(kustomization_yaml_unstamped_file, kustomization_yaml)
        _stamp_file(ctx, kustomization_yaml_unstamped_file, kustomization_yaml_file)
    else:
        ctx.actions.write(kustomization_yaml_file, kustomization_yaml)

    transitive_runfiles = []
    resolver_part = ""
    if ctx.attr.images:
        resolver_part += " | {resolver} ".format(resolver = ctx.executable._resolver.path)
        tmpfiles.append(ctx.executable._resolver)
        for img in ctx.attr.images:
            kpi = img[K8sPushInfo]
            regrepo = kpi.registry + "/" + kpi.repository
            if "{" in regrepo:
                regrepo = stamp(ctx, regrepo, tmpfiles, ctx.attr.name + regrepo.replace("/", "_"))

            resolver_part += " --image {}={}@$(cat {})".format(str(kpi.image_label).replace("@@//", "//"), regrepo, kpi.digestfile.path)

            tmpfiles.append(kpi.digestfile)
            transitive_runfiles.append(img[DefaultInfo].default_runfiles)

    template_part = ""
    if ctx.attr.substitutions or ctx.attr.deps:
        template_part += "| {} --stamp_info_file={} ".format(ctx.executable._template_engine.path, ctx.file._info_file.path)
        tmpfiles.append(ctx.executable._template_engine)
        tmpfiles.append(ctx.file._info_file)
        for k in ctx.attr.substitutions:
            template_part += "--variable=%s=%s " % (k, ctx.attr.substitutions[k])
        if ctx.attr.start_tag:
            template_part += "--start_tag=%s " % ctx.attr.start_tag
        if ctx.attr.end_tag:
            template_part += "--end_tag=%s " % ctx.attr.end_tag
        d = {
            str(ctx.attr.deps[i].label): ctx.files.deps[i].path
            for i in range(0, len(ctx.attr.deps))
        }
        template_part += " ".join(["--imports=%s=%s" % (k, d[k]) for k in d])
        template_part += " "
        template_part += " ".join([
            "--imports=%s=%s" % (k, d[str(ctx.label.relative(ctx.attr.deps_aliases[k]))])
            for k in ctx.attr.deps_aliases
        ])

        # Image name substitutions
        if ctx.attr.images:
            for img in ctx.attr.images:
                kpi = img[K8sPushInfo]
                regrepo = kpi.registry + "/" + kpi.repository
                if "{" in regrepo:
                    regrepo = stamp(ctx, regrepo, tmpfiles, ctx.attr.name + regrepo.replace("/", "_"))
                template_part += " --variable={}={}@$(cat {})".format(str(kpi.image_label).replace("@@//", "//"), regrepo, kpi.digestfile.path)

                # Image digest
                template_part += " --variable={}=$(cat {} | cut -d ':' -f 2)".format(str(kpi.image_label) + ".digest", kpi.digestfile.path)
                template_part += " --variable={}=$(cat {} | cut -c 8-17)".format(str(kpi.image_label) + ".short-digest", kpi.digestfile.path)
                if str(kpi.image_label).startswith("@//"):
                    # Bazel 6 add a @ prefix to the image label
                    label = str(kpi.image_label)[1:]
                    template_part += " --variable={}=$(cat {} | cut -d ':' -f 2)".format(str(label) + ".digest", kpi.digestfile.path)
                    template_part += " --variable={}=$(cat {} | cut -c 8-17)".format(str(label) + ".short-digest", kpi.digestfile.path)

                if hasattr(kpi, "legacy_image_name"):
                    template_part += " --variable={}={}@$(cat {})".format(kpi.legacy_image_name, regrepo, kpi.digestfile.path)

        template_part += " "

    kustomize_executable = ctx.toolchains["@rules_gitops//kustomize:toolchain_type"].kustomizeinfo.executable

    script = ctx.actions.declare_file("%s-kustomize" % ctx.label.name)
    script_content = _script_template.format(
        kustomize = kustomize_executable.path,
        kustomize_dir = root,
        resolver_part = resolver_part,
        template_part = template_part,
        out = ctx.outputs.yaml.path,
    )
    ctx.actions.write(script, script_content, is_executable = True)

    ctx.actions.run(
        outputs = [ctx.outputs.yaml],
        inputs = ctx.files.manifests + ctx.files.configmaps_srcs + ctx.files.secrets_srcs + ctx.files.configurations + [kustomization_yaml_file] + tmpfiles + ctx.files.patches + ctx.files.deps,
        executable = script,
        mnemonic = "Kustomize",
        tools = [kustomize_executable],
    )

    runfiles = ctx.runfiles(files = ctx.files.deps).merge_all(transitive_runfiles)

    transitive_files = [m[DefaultInfo].files for m in ctx.attr.manifests if KustomizeInfo in m]
    transitive_files += [obj[DefaultInfo].files for obj in ctx.attr.objects]

    transitive_image_pushes = [m[KustomizeInfo].image_pushes for m in ctx.attr.manifests if KustomizeInfo in m]
    transitive_image_pushes += [obj[KustomizeInfo].image_pushes for obj in ctx.attr.objects]

    return [
        DefaultInfo(
            files = depset(
                [ctx.outputs.yaml],
                transitive = transitive_files,
            ),
            runfiles = runfiles,
        ),
        KustomizeInfo(
            image_pushes = depset(
                ctx.attr.images,
                transitive = transitive_image_pushes,
            ),
        ),
    ]

kustomization = rule(
    implementation = _kustomization_impl,
    attrs = {
        "configmaps_srcs": attr.label_list(allow_files = True),
        "secrets_srcs": attr.label_list(allow_files = True),
        "deps_aliases": attr.string_dict(default = {}),
        "disable_name_suffix_hash": attr.bool(default = True),
        "end_tag": attr.string(default = "}}"),
        "images": attr.label_list(doc = "a list of image pushes used in manifests", providers = [K8sPushInfo]),
        "manifests": attr.label_list(allow_files = True),
        "name_prefix": attr.string(),
        "name_suffix": attr.string(),
        "namespace": attr.string(),
        "objects": attr.label_list(doc = "a list of dependent kustomize objects", providers = [KustomizeInfo]),
        "patches": attr.label_list(allow_files = True),
        "image_name_patches": attr.string_dict(default = {}, doc = "set new names for selected images"),
        "image_tag_patches": attr.string_dict(default = {}, doc = "set new tags for selected images"),
        "start_tag": attr.string(default = "{{"),
        "substitutions": attr.string_dict(default = {}),
        "deps": attr.label_list(default = [], allow_files = True),
        "configurations": attr.label_list(allow_files = True),
        "common_labels": attr.string_dict(default = {}),
        "common_annotations": attr.string_dict(default = {}),
        "_build_user_value": attr.label(
            default = Label("//stamper:build_user_value.txt"),
            allow_single_file = True,
        ),
        "_info_file": attr.label(
            default = Label("//stamper:more_stable_status.txt"),
            allow_single_file = True,
        ),
        "_resolver": attr.label(
            default = Label("//resolver:resolver"),
            cfg = "exec",
            executable = True,
        ),
        "_stamper": attr.label(
            default = Label("//stamper:stamper"),
            cfg = "exec",
            executable = True,
            allow_files = True,
        ),
        "_template_engine": attr.label(
            default = Label("//templating:fast_template_engine"),
            executable = True,
            cfg = "exec",
        ),
    },
    outputs = {
        "yaml": "%{name}.yaml",
    },
    toolchains = ["@rules_gitops//kustomize:toolchain_type"],
)
