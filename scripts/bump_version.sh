#!/bin/bash
# Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
set -e
source "$(dirname $(realpath $0))/common.sh"
current_folder="$(dirname $(realpath $0))"

command -v yq >/dev/null 2>&1 || die "yq is required but not found on PATH"

DEFAULT_ECR_URL="public.ecr.aws/aws-nitro-enclaves"
if [[ -f "$ECR_CONFIG_FILE_PATH" ]]; then
  source "$ECR_CONFIG_FILE_PATH"
fi
ECR_URL="${ECR_URL:-$DEFAULT_ECR_URL}"

say "Bumping version references to $RELEASE (registry: $ECR_URL)..."

# aws-nitro-enclaves-k8s-ds.yaml — update the full image (registry + tag)
# select(.kind == "DaemonSet") scopes the update to the DaemonSet document only,
# preventing yq from adding spurious spec fields to the Namespace document
yq -i "(select(.kind == \"DaemonSet\") | .spec.template.spec.containers[] | select(.name == \"aws-nitro-enclaves-k8s-dp\")).image = \"$ECR_URL/$REPOSITORY_NAME:$RELEASE\"" \
  "$TOP_DIR/aws-nitro-enclaves-k8s-ds.yaml"
say "  Updated aws-nitro-enclaves-k8s-ds.yaml"

# helm/Chart.yaml — update version and appVersion
yq -i ".version = \"$RELEASE\" | .appVersion = \"$RELEASE\"" "$TOP_DIR/helm/Chart.yaml"
say "  Updated helm/Chart.yaml"

# helm/values.yaml — update image repository and tag
yq -i ".awsNitroEnclavesK8SDaemonset.awsNitroEnclavesK8SDp.image.repository = \"$ECR_URL/$REPOSITORY_NAME\" | .awsNitroEnclavesK8SDaemonset.awsNitroEnclavesK8SDp.image.tag = \"$RELEASE\"" \
  "$TOP_DIR/helm/values.yaml"
say "  Updated helm/values.yaml"

# validate all versions are in sync
$current_folder/validate_artifacts_versions.sh

say "All version references updated to $RELEASE"
