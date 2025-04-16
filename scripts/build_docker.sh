#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
source "$(dirname $(realpath $0))/common.sh"

build_docker_image() {
    local arch=$1
    docker build --target device_plugin --platform linux/$arch -t $IMAGE-$arch $TOP_DIR -f $TOP_DIR/container/Dockerfile
}

docker build --target builder -t $BUILDER_IMAGE $TOP_DIR -f $TOP_DIR/container/Dockerfile ||
    die "Failed to build generic builder image"
arch=x86_64 && build_docker_image ${arch} || die "Failed to build ${arch} image"
arch=aarch64 && build_docker_image ${arch} || die "Failed to build ${arch} image"
