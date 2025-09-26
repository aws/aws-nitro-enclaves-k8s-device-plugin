#!/bin/bash
# Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
source "$(dirname $(realpath $0))/common.sh"

main() {
  helm_login

  aws ecr-public --region $ECR_REGION describe-repositories \
    --repository-names "charts/$REPOSITORY_NAME" >/dev/null ||
    die "There is no repository named $REPOSITORY_NAME in" \
      "$ECR_REGION region."

  is_a_public_ecr_registry && {
    confirm "You are about to push a $RELEASE Helm chart on a public repository." \
      "Are you sure you want to continue?"
  }
  say "Pushing $HELM_CHART to $ECR_HELM_URL..."
  helm push aws-nitro-enclaves-k8s-device-plugin-chart-$RELEASE.tgz oci://$ECR_HELM_URL ||
    die "Failed to push $HELM_CHART to $ECR_HELM_URL."
}

main
