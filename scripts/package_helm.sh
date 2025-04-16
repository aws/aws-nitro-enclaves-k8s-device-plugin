#!/bin/bash
# Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
source "$(dirname $(realpath $0))/common.sh"

helm lint $TOP_DIR/helm && helm package $TOP_DIR/helm ||
    die "Helm package lint failed"

# assert that packaged file is located in directory
# its best practice to manage helm version and app relase version independent from each other
# VERSION is sourced from packed RELEASE veriable and HELM versions are based on Chart.yaml values
if [[ ! -f $TOP_DIR/aws-nitro-enclaves-k8s-device-plugin-$VERSION.tgz ]]; then
    die "Packaged file not found in $TOP_DIR directory"
fi

# change name of standard HELM archive to explicitly state that it is a packaged chart
mv aws-nitro-enclaves-k8s-device-plugin-$VERSION.tgz $HELM_CHART
