#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# --load is required when using Docker BuildKit to make the image available
# in the local image store. Override with BUILDX_LOAD="" when using engines
# that store images locally by default (e.g. Podman).
BUILDX_LOAD="${BUILDX_LOAD:---load}"

# Optional cache directives (e.g. --cache-from type=gha --cache-to type=gha,mode=max
# in CI, or --cache-from type=local,src=/tmp/cache locally). Empty by default.
BUILDX_CACHE="${BUILDX_CACHE:-}"

source "$(dirname $(realpath $0))/common.sh"

build_docker_image() {
    local arch=$1
    docker buildx build $BUILDX_LOAD $BUILDX_CACHE --target device_plugin --platform linux/$arch -t $IMAGE-$arch $TOP_DIR -f $TOP_DIR/container/Dockerfile
}

# Build both architectures in parallel; BuildKit caches the shared builder stage.
build_docker_image x86_64 &
pid_x86=$!
build_docker_image aarch64 &
pid_arm=$!
wait $pid_x86 || die "Failed to build x86_64 image"
wait $pid_arm || die "Failed to build aarch64 image"
