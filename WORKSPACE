workspace(name = "com_github_istio_mixer")

git_repository(
    name = "io_bazel_rules_go",
    commit = "87cdda3fc0fd65c63ef0316533be03ea4956f809",  # April 7 2017 (0.4.2)
    remote = "https://github.com/bazelbuild/rules_go.git",
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "new_go_repository")

go_repositories()

git_repository(
    name = "org_pubref_rules_protobuf",
    commit = "9ede1dbc38f0b89ae6cd8e206a22dd93cc1d5637",  # Mar 31 2017 (gogo* support)
    remote = "https://github.com/pubref/rules_protobuf",
)

load("@org_pubref_rules_protobuf//protobuf:rules.bzl", "proto_repositories")

proto_repositories()

load("@org_pubref_rules_protobuf//gogo:rules.bzl", "gogo_proto_repositories")
load("@org_pubref_rules_protobuf//cpp:rules.bzl", "cpp_proto_repositories")

cpp_proto_repositories()

gogo_proto_repositories()

new_go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b4da8ac2ed8007337bc5eb5007998",  # Jan 26, 2016 (no releases)
    importpath = "github.com/golang/glog",
)

new_go_repository(
    name = "com_github_ghodss_yaml",
    commit = "04f313413ffd65ce25f2541bfd2b2ceec5c0908c",  # Dec 6, 2016 (no releases)
    importpath = "github.com/ghodss/yaml",
)

new_go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "14227de293ca979cf205cd88769fe71ed96a97e2",  # Jan 24, 2017 (no releases)
    importpath = "gopkg.in/yaml.v2",
)

new_go_repository(
    name = "com_github_golang_protobuf",
    commit = "8ee79997227bf9b34611aee7946ae64735e6fd93",  # Nov 16, 2016 (no releases)
    importpath = "github.com/golang/protobuf",
)

GOOGLEAPIS_BUILD_FILE = """
package(default_visibility = ["//visibility:public"])

load("@io_bazel_rules_go//go:def.bzl", "go_prefix")
go_prefix("github.com/googleapis/googleapis")

load("@org_pubref_rules_protobuf//gogo:rules.bzl", "gogoslick_proto_library")

gogoslick_proto_library(
    name = "google/rpc",
    protos = [
        "google/rpc/code.proto",
        "google/rpc/error_details.proto",
        "google/rpc/status.proto",
    ],
    importmap = {
        "google/protobuf/any.proto": "github.com/gogo/protobuf/types",
        "google/protobuf/duration.proto": "github.com/gogo/protobuf/types",
    },
    imports = [
        "../../external/com_github_google_protobuf/src",
    ],
    inputs = [
        "@com_github_google_protobuf//:well_known_protos",
    ],
    deps = [
        "@com_github_gogo_protobuf//types:go_default_library",
    ],
    verbose = 0,
)

load("@org_pubref_rules_protobuf//cpp:rules.bzl", "cc_proto_library")

cc_proto_library(
    name = "cc_status_proto",
    protos = [
        "google/rpc/status.proto",
    ],
    imports = [
        "../../external/com_github_google_protobuf/src",
    ],
    verbose = 0,
)

filegroup(
    name = "status_proto",
    srcs = [ "google/rpc/status.proto" ],
)

filegroup(
    name = "code_proto",
    srcs = [ "google/rpc/code.proto" ],
)
"""

new_git_repository(
    name = "com_github_googleapis_googleapis",
    build_file_content = GOOGLEAPIS_BUILD_FILE,
    commit = "13ac2436c5e3d568bd0e938f6ed58b77a48aba15",  # Oct 21, 2016 (only release pre-dates sha)
    remote = "https://github.com/googleapis/googleapis.git",
)

new_go_repository(
    name = "org_golang_google_grpc",
    commit = "cdee119ee21e61eef7093a41ba148fa83585e143",  # Mar 14, 2017 (v1.2.0)
    importpath = "google.golang.org/grpc",
)

new_go_repository(
    name = "com_github_spf13_cobra",
    commit = "35136c09d8da66b901337c6e86fd8e88a1a255bd",  # Jan 30, 2017 (no releases)
    importpath = "github.com/spf13/cobra",
)

new_go_repository(
    name = "com_github_spf13_pflag",
    commit = "9ff6c6923cfffbcd502984b8e0c80539a94968b7",  # Jan 30, 2017 (no releases)
    importpath = "github.com/spf13/pflag",
)

new_go_repository(
    name = "com_github_cpuguy83_go_md2man",
    commit = "648eed146d3f3beacb64063cd0daae908015eebd",  # Mar 19, 2017 (no releases)
    importpath = "github.com/cpuguy83/go-md2man",
)

new_go_repository(
    name = "com_github_russross_blackfriday",
    commit = "35eb537633d9950afc8ae7bdf0edb6134584e9fc",  # Mar 19, 2017 (no releases)
    importpath = "github.com/russross/blackfriday",
)

new_go_repository(
    name = "com_github_shurcooL_sanitized_anchor_name",
    commit = "10ef21a441db47d8b13ebcc5fd2310f636973c77",  # Mar 19, 2017 (no releases)
    importpath = "github.com/shurcooL/sanitized_anchor_name",
)

new_go_repository(
    name = "com_github_hashicorp_go_multierror",
    commit = "ed905158d87462226a13fe39ddf685ea65f1c11f",  # Dec 16, 2016 (no releases)
    importpath = "github.com/hashicorp/go-multierror",
)

new_go_repository(
    name = "com_github_hashicorp_errwrap",
    commit = "7554cd9344cec97297fa6649b055a8c98c2a1e55",  # Oct 27, 2014 (no releases)
    importpath = "github.com/hashicorp/errwrap",
)

new_go_repository(
    name = "com_github_opentracing_opentracing_go",
    commit = "0c3154a3c2ce79d3271985848659870599dfb77c",  # Sep 26, 2016 (v1.0.0)
    importpath = "github.com/opentracing/opentracing-go",
)

new_go_repository(
    name = "com_github_opentracing_basictracer",
    commit = "1b32af207119a14b1b231d451df3ed04a72efebf",  # Sep 29, 2016 (no releases)
    importpath = "github.com/opentracing/basictracer-go",
)

load("//:repositories.bzl", "new_git_or_local_repository")

new_git_or_local_repository(
    name = "com_github_istio_api",
    build_file = "BUILD.api",
    commit = "86f8a709b94340ad381ea8ac2778d469b85bd928",  # Apr 14, 2017 (no releases)
    path = "../api",
    remote = "https://github.com/istio/api.git",
    # Change this to True to use ../api directory
    use_local = False,
)

new_http_archive(
    name = "docker_ubuntu",
    build_file = "BUILD.ubuntu",
    sha256 = "2c63dd81d714b825acd1cb3629c57d6ee733645479d0fcdf645203c2c35924c5",
    type = "zip",
    url = "https://codeload.github.com/tianon/docker-brew-ubuntu-core/zip/b6f1fe19228e5b6b7aed98dcba02f18088282f90",
)

DEBUG_BASE_IMAGE_SHA = "3f57ae2aceef79e4000fb07ec850bbf4bce811e6f81dc8cfd970e16cdf33e622"

# See github.com/istio/manager/blob/master/docker/debug/build-and-publish-debug-image.sh
# for instructions on how to re-build and publish this base image layer.
http_file(
    name = "ubuntu_xenial_debug",
    sha256 = DEBUG_BASE_IMAGE_SHA,
    url = "https://storage.googleapis.com/istio-build/manager/ubuntu_xenial_debug-" + DEBUG_BASE_IMAGE_SHA + ".tar.gz",
)

new_go_repository(
    name = "com_github_prometheus_client_golang",
    commit = "c5b7fccd204277076155f10851dad72b76a49317",  # Aug 17, 2016 (v0.8.0)
    importpath = "github.com/prometheus/client_golang",
)

new_go_repository(
    name = "com_github_prometheus_common",
    commit = "dd2f054febf4a6c00f2343686efb775948a8bff4",  # Jan 8, 2017 (no releases)
    importpath = "github.com/prometheus/common",
)

new_go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    commit = "c12348ce28de40eed0136aa2b644d0ee0650e56c",  # Apr 24, 2016 (v1.0.0)
    importpath = "github.com/matttproud/golang_protobuf_extensions",
)

new_go_repository(
    name = "com_github_prometheus_procfs",
    commit = "1878d9fbb537119d24b21ca07effd591627cd160",  # Jan 28, 2017 (no releases)
    importpath = "github.com/prometheus/procfs",
)

new_go_repository(
    name = "com_github_beorn7_perks",
    commit = "4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9",  # Aug 4, 2016 (no releases)
    importpath = "github.com/beorn7/perks",
)

new_go_repository(
    name = "com_github_prometheus_client_model",
    commit = "fa8ad6fec33561be4280a8f0514318c79d7f6cb6",  # Feb 12, 2015 (only release too old)
    importpath = "github.com/prometheus/client_model",
)

new_go_repository(
    name = "com_github_cactus_statsd_client",
    commit = "91c326c3f7bd20f0226d3d1c289dd9f8ce28d33d",  # release 3.1.0, 5/30/2016
    importpath = "github.com/cactus/go-statsd-client",
)

new_go_repository(
    name = "com_github_redis_client",
    commit = "1ac54a28f5934ea5e08f588647e734aba2383cb8",  # Jan 28, 2017 (no releases)
    importpath = "github.com/mediocregopher/radix.v2",
)

new_go_repository(
    name = "com_github_mini_redis",
    commit = "e9169f14d501184b6cc94e270e5a93e4bab203d7",  # release 2.0.0, 4/15/2017
    importpath = "github.com/alicebob/miniredis",
)

new_go_repository(
    name = "com_github_bsm_redeo",
    commit = "1ce09fc76693fb3c1ca9b529c66f38920beb6fb8",  # Aug 17, 2016 (no releases)
    importpath = "github.com/bsm/redeo",
)

new_go_repository(
    name = "com_github_emicklei_go_restful",
    commit = "1d2cd46e1ac9ce6036f014a33a3706f989578de6",  # March 17, 2017 (no releases)
    importpath = "github.com/emicklei/go-restful",
)
