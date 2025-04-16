#!/bin/bash
# Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
source "$(dirname $(realpath $0))/common.sh"

# extract version of kubernetes manifest
k8s_manifest=$TOP_DIR/aws-nitro-enclaves-k8s-ds.yaml
k8s_version=$(yq '.spec.template.spec.containers[]?.image' "$k8s_manifest" | grep -o '[^:]*$')

# extract version of helm chart, should be based on k8s manifest
helm_chart=$TOP_DIR/helm/values.yaml
helm_version=$(yq '.awsNitroEnclavesK8SDaemonset.awsNitroEnclavesK8SDp.image.tag' $helm_chart)

echo "Release: $RELEASE"
echo "Kubernetes Manifest: $k8s_version"
echo "Helm Chart: $helm_version"

if [ $RELEASE != $k8s_version ] || [ $k8s_version != $helm_version ]; then
    die "Versions in release $RELEASE are not in sync"
fi
