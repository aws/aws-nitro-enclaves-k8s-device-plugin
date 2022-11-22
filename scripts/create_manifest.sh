#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

source "$(dirname $(realpath $0))/common.sh"

main() {
  ecr_login

  docker manifest create --amend $ECR_URL/$REPOSITORY_NAME \
    $ECR_URL/$REPOSITORY_NAME:$RELEASE-x86_64 \
    $ECR_URL/$REPOSITORY_NAME:$RELEASE-aarch64 || \
    die "Cannot create manifest for multiarch image." \
      " Please ensure that both x86_64 and aarch64 images" \
      " already exist in the repository."

  docker manifest inspect $ECR_URL/$IMAGE
  
  is_a_public_ecr_registry && {
    confirm "You are about to make changes on a" \
      " publicly available manifest. Are you sure want to continue? (yes/no)"
  }

  docker manifest push $ECR_URL/$REPOSITORY_NAME:latest
}

main
