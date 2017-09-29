load("@io_bazel_rules_go//go:def.bzl", "go_repository")

def mixer_adapter_repositories():

    native.bind(
        name = "protoc",
        actual = "@com_github_google_protobuf//:protoc",
    )

    native.bind(
        name = "protocol_compiler",
        actual = "@com_github_google_protobuf//:protoc",
    )

    go_repository(
        name = "com_github_golang_glog",
        commit = "23def4e6c14b4da8ac2ed8007337bc5eb5007998",  # Jan 26, 2016 (no releases)
        importpath = "github.com/golang/glog",
    )

    go_repository(
        name = "com_github_golang_protobuf",
        commit = "8ee79997227bf9b34611aee7946ae64735e6fd93",  # Nov 16, 2016 (match pubref dep)
        importpath = "github.com/golang/protobuf",
    )

    go_repository(
        name = "com_github_gogo_protobuf",
        commit = "100ba4e885062801d56799d78530b73b178a78f3",  # Mar 7, 2017 (match pubref dep)
        importpath = "github.com/gogo/protobuf",
    )

    go_repository(
        name = "org_golang_google_grpc",
        commit = "d2e1b51f33ff8c5e4a15560ff049d200e83726c5",  # April 28, 2017 (v1.3.0)
        importpath = "google.golang.org/grpc",
    )

    go_repository(
        name = "org_golang_x_text",
        build_file_name = "BUILD.bazel",
        commit = "f4b4367115ec2de254587813edaa901bc1c723a8",  # Mar 31, 2017 (no releases)
        importpath = "golang.org/x/text",
    )
