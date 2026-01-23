# Copyright 2026 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

def stamp(ctx, string, files, tmpfilename):
    """Stamp provided string replacing placeholders like {BUILD_USER}.

    Uses an optimization shortcut for BUILD_USER

    Returns:
        a string suitable for inclusion into bash script.
    """
    deps = []

    if "{BUILD_USER}" in string and "{" not in string.format(BUILD_USER = ""):
        # shortcut for only {BUILD_USER} in placeholders
        string = string.format(
            BUILD_USER = "$(cat %s)" % ctx.file._build_user_value.path,
        )

        deps.append(ctx.files._build_user_value[0])
        return string, deps

    stamps = [ctx.file._info_file]
    stamp_args = [
        "--stamp-info-file=%s" % sf.path
        for sf in stamps
    ]
    tmp_out_file = ctx.actions.declare_file(tmpfilename)

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
    string = "$(cat {})".format(tmp_out_file.path)
    deps.append(tmp_out_file)
    return string, deps

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
            default = Label("//stamper:more_stable_status.txt"),
            allow_single_file = True,
        ),
        "_stamper": attr.label(
            default = Label("//stamper:stamper"),
            cfg = "exec",
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
