#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Build the multi-arch device plugin container images.
#
# Environment variables (exported by the caller, e.g. release.sh or CI):
#   BUILDX_LOAD  Passed to `docker buildx build`. Defaults to `--load` to make
#                the image available in the local image store. Set to empty
#                when using engines that already store images locally (Podman).
#   BUILDX_CACHE Optional cache directives, e.g.
#                `--cache-from type=gha --cache-to type=gha,mode=max` in CI.
source "$(dirname $(realpath $0))/common.sh"

build_docker_image() {
    local arch=$1
    docker buildx build ${BUILDX_LOAD:---load} ${BUILDX_CACHE:-} --target device_plugin --platform linux/$arch -t $IMAGE-$arch $TOP_DIR -f $TOP_DIR/container/Dockerfile
}

# Build both architectures in parallel; BuildKit caches the shared builder stage.
build_docker_image x86_64 &
pid_x86=$!
build_docker_image aarch64 &
pid_arm=$!
wait $pid_x86 || die "Failed to build x86_64 image"
wait $pid_arm || die "Failed to build aarch64 image"
