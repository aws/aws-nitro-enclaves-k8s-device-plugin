#!/bin/bash
# Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
set -e
source "$(dirname $(realpath $0))/common.sh"
current_folder="$(dirname $(realpath $0))"

# version of helm charts are based on /helm/Chart.yaml
# before packaging and publishing validate that the RELEASE version, manifest.yaml
# and helm chart version are in sync and pointig to the new multich arch docker manifest
$current_folder/validate_artifacts_versions.sh

# build and upload docker artifacts
# version for docker artifacts are based on RELEASE file
$current_folder/build_docker.sh
$current_folder/push_docker.sh
$current_folder/create_manifest_docker.sh

# build and upload helm artifacts
$current_folder/package_helm.sh
$current_folder/push_helm.sh
