#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

source "$(dirname $(realpath $0))/common.sh"

docker build --target builder -t $BUILDER_IMAGE $TOP_DIR -f $TOP_DIR/container/Dockerfile
docker build --target device_plugin -t $IMAGE $TOP_DIR -f $TOP_DIR/container/Dockerfile
