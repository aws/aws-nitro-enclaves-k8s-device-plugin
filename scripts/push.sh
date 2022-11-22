#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

source "$(dirname $(realpath $0))/common.sh"

main() {
  ecr_login

  aws ecr --region $ECR_REGION describe-repositories \
    --repository-names "$REPOSITORY_NAME" > /dev/null || \
    die "There is no repository named $REPOSITORY_NAME in" \
    "$ECR_REGION region."

  is_a_public_ecr_registry && {
    confirm "You are about to make changes on a public repository." \
            " Are you sure want to continue?"
  }

  docker tag $IMAGE $ECR_URL/$IMAGE
  say "Pushing $IMAGE to $ECR_URL..."
  docker push $ECR_URL/$IMAGE
}

main
