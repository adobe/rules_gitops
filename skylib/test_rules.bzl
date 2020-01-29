# Copyright 2020 Adobe. All rights reserved.
# This file is licensed to you under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License. You may obtain a copy
# of the License at http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
# OF ANY KIND, either express or implied. See the License for the specific language
# governing permissions and limitations under the License.

def _file_compare_test_impl(ctx):
    """check that a file has a given content."""
    exe = ctx.outputs.executable
    file_ = ctx.file.file
    expected_ = ctx.file.expected
    ctx.actions.write(
        output = exe,
        content = "diff -u %s %s" % (expected_.short_path, file_.short_path),
        is_executable = True,
    )
    return [
        DefaultInfo(runfiles = ctx.runfiles([exe, expected_, file_])),
    ]

file_compare_test = rule(
    attrs = {
        "expected": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
        "file": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
    },
    executable = True,
    test = True,
    implementation = _file_compare_test_impl,
)
