"""Utility functions for golden testing."""

load("@bazel_lib//lib:write_source_files.bzl", "write_source_file")

def golden_test(name, in_file, extension = ""):
    write_source_file(
        name = name,
        in_file = in_file,
        out_file = "goldens/{}.golden{}".format(in_file, extension),
        diff_test = True,
        testonly = True,
    )
