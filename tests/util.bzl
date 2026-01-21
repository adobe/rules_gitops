load("@bazel_lib//lib:write_source_files.bzl", "write_source_file")

def file_compare_test(name, expected, file):
    write_source_file(
        name = name,
        in_file = file,
        out_file = expected,
        diff_test = True,
        testonly = True
    )