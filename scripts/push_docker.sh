#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
source "$(dirname $(realpath $0))/common.sh"

tag_and_push_docker_image() {
  local arch=$1

  docker tag $IMAGE-$arch $ECR_URL/$IMAGE-$arch
  say "Pushing $IMAGE-$arch to $ECR_URL..."
  docker push $ECR_URL/$IMAGE-$arch
}

main() {
  ecr_login

  aws ecr-public --region $ECR_REGION describe-repositories \
    --repository-names "$REPOSITORY_NAME" >/dev/null ||
    die "There is no repository named $REPOSITORY_NAME in" \
      "$ECR_REGION region."

  is_a_public_ecr_registry && {
    confirm "You are about to push $RELEASE docker images on a public repository." \
      "Are you sure you want to continue?"
  }

  arch=x86_64 && tag_and_push_docker_image ${arch} || die "Failed to push $arch docker image"
  arch=aarch64 && tag_and_push_docker_image ${arch} || die "Failed to push $arch docker image"
}

main
