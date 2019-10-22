# Copyright (c) 2018 Lotus Load
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load("@io_bazel_rules_docker//container:container.bzl", "container_bundle")
load("@io_bazel_rules_docker//contrib:push-all.bzl", "docker_push")

def app_image(name, binary, repository, **kwargs):
    go_image(
        name = "image",
        binary = binary,
        **kwargs)

    container_bundle(
        name = "bundle",
        images = {
            "$(DOCKER_REGISTRY)/%s:{STABLE_GIT_COMMIT}" % repository: ":image",
            "$(DOCKER_REGISTRY)/%s:{STABLE_GIT_COMMIT_FULL}" % repository: ":image",
        },
    )

    docker_push(
        name = "push",
        bundle = ":bundle",
        **kwargs)

