# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

load(
    "@io_bazel_rules_docker//skylib:path.bzl",
    "runfile",
)

def file_path(ctx, f, in_runtime):
    if in_runtime:
        return "${RUNFILES}/%s" % runfile(ctx, f)
    else:
        return f.path

def stamp(ctx, string, files, tmpfilename, in_runtime = False):
    """
    Stamp provided string replacing placeholders like {BUILD_USER}.
    Uses an optimization shortcut for BUILD_USER
    Returns a string suitable for inclusion into bash script.
    """
    if "{BUILD_USER}" in string and "{" not in string.format(BUILD_USER = ""):
        # shortcut for only {BUILD_USER} in placeholders
        string = string.format(
            BUILD_USER = "$(cat %s)" % file_path(ctx, ctx.file._build_user_value, in_runtime),
        )

        files.append(ctx.files._build_user_value[0])
        return string

    stamps = [ctx.file._info_file]
    stamp_args = [
        "--stamp-info-file=%s" % sf.path
        for sf in stamps
    ]
    tmp_out_file = ctx.actions.declare_file(tmpfilename)
    files.append(tmp_out_file)
    ctx.actions.run(
        executable = ctx.executable._stamper,
        arguments = [
            "--format=%s" % string,
            "--output=%s" % tmp_out_file.path,
        ] + stamp_args,
        inputs = stamps,
        outputs = [tmp_out_file],
        mnemonic = "Stamp",
        tools = [ctx.executable._stamper],
    )
    string = "$(cat {})".format(file_path(ctx, tmp_out_file, in_runtime))
    return string

def _stamp_value_impl(ctx):
    stamps = [ctx.file._info_file]
    stamp_args = [
        "--stamp-info-file=%s" % sf.path
        for sf in stamps
    ]
    ctx.actions.run(
        executable = ctx.executable._stamper,
        arguments = [
            "--format=%s" % ctx.attr.str,
            "--output=%s" % ctx.outputs.out.path,
        ] + stamp_args,
        inputs = stamps,
        outputs = [ctx.outputs.out],
        mnemonic = "Stamp",
        tools = [ctx.executable._stamper],
    )

stamp_value = rule(
    implementation = _stamp_value_impl,
    attrs = {
        "str": attr.string(default = "{BUILD_USER}"),
        "_info_file": attr.label(
            default = Label("//skylib:more_stable_status.txt"),
            allow_single_file = True,
        ),
        "_stamper": attr.label(
            default = Label("//stamper:stamper"),
            cfg = "host",
            executable = True,
            allow_files = True,
        ),
    },
    outputs = {
        "out": "%{name}.txt",
    },
)

def _more_stable_status_impl(ctx):
    v = " ".join(["-e ^" + var for var in ctx.attr.vars])
    ctx.actions.run_shell(
        inputs = [ctx.info_file],
        outputs = [ctx.outputs.out],
        progress_message = "Filtering stable status file",
        command = "grep {} {} >{}".format(v, ctx.info_file.path, ctx.outputs.out.path),
    )

# Generate reduced more stable version of stable-status.txt
# Limited number of rows is extracted now to make it cacheable for CI/CD
more_stable_status = rule(
    attrs = {
        "vars": attr.string_list(
            mandatory = True,
            doc = "Variables to extract from stable_status.txt",
        ),
    },
    outputs = {
        "out": "%{name}.txt",
    },
    implementation = _more_stable_status_impl,
)
