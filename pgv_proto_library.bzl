load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//proto:compiler.bzl", "go_proto_compiler")

def pgv_go_proto_library(name, proto = None, deps = [], **kwargs):
    go_proto_compiler(
        name = "pgv_plugin_go",
        suffix = ".pb.validate.go",
        valid_archive = False,
        plugin = "@com_lyft_protoc_gen_validate//:protoc-gen-validate",
        options = ["lang=go"],
    )

    go_proto_library(
        name = name,
        proto = proto,
        deps = ["@com_lyft_protoc_gen_validate//validate:go_default_library"] + deps,
        compilers = [
            "@io_bazel_rules_go//proto:go_proto",
            "pgv_plugin_go",
        ],
        visibility = ["//visibility:public"],
        **kwargs
    )

