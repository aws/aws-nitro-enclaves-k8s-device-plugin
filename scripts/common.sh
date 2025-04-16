#!/bin/bash
# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

readonly SUCCESS=0
readonly FAILURE=255

readonly SCRIPTS_DIR=$(dirname $(realpath $0))
readonly TOP_DIR=$(cd $SCRIPTS_DIR/.. && pwd)
readonly ECR_CONFIG_FILE_PATH="$SCRIPTS_DIR/.ecr.uri"
readonly RELEASE_FILE="RELEASE"

readonly BUILDER_IMAGE=ne-k8s-device-plugin-build:latest
readonly REPOSITORY_NAME=aws-nitro-enclaves-k8s-device-plugin
readonly RELEASE=$(cat $TOP_DIR/$RELEASE_FILE)
readonly TAG=$RELEASE-$(arch)
readonly IMAGE=$REPOSITORY_NAME:$TAG

say() {
  echo "$@"
}

die() {
  say "[ERROR] $@"
  exit $FAILURE
}

[[ -f $TOP_DIR/$RELEASE_FILE ]] || \
  die "Cannot find $RELEASE_FILE file in $TOP_DIR directory."

_set_config_item() {
  local var=$1; shift
  local prompt="$@"

  local value=""
  while [[ $value = "" ]];
  do
      printf "$prompt"
      read value
  done

  echo "$var=$value" >> "$ECR_CONFIG_FILE_PATH"
}

_load_ecr_config() {
  [[ -f $ECR_CONFIG_FILE_PATH ]] || {
      printf "No configuration found!\n"
      _set_config_item ECR_URL "Please enter an ECR URL:"
      _set_config_item ECR_REGION "Please enter AWS region of the ECR repository:"
  }

  source "$ECR_CONFIG_FILE_PATH"
  [[ -z "$ECR_URL" || -z "$ECR_REGION" ]] && {
      say "$(basename $ECR_CONFIG_FILE_PATH) seems corrupted. Try using" \
      "'rm -f $ECR_CONFIG_FILE_PATH' to remove this configuration."
      exit 1
  }

  return 0
}

_ecr_login() {
  is_a_public_ecr_registry

  if [[ $? -eq $SUCCESS ]]; then
      aws ecr-public get-login-password --region "$ECR_REGION" | docker login --username AWS --password-stdin $ECR_URL
  else
      aws ecr get-login-password --region "$ECR_REGION" | docker login --username AWS --password-stdin $ECR_URL
  fi
}

# Loads configuration and logs in to a registry.
#
ecr_login() {
    _load_ecr_config || die "Error while loading configuration file!"
    say "Using ECR registry url: $ECR_URL. (region: $ECR_REGION)."
    _ecr_login || die "Failed to log in to the ECR registry."
}

# Check if the current ECR URL is a public one or not.
# 
is_a_public_ecr_registry() {
  [[ "$ECR_URL" =~ ^public.ecr.aws* ]] && { return $SUCCESS; }
  return $FAILURE
}

_helm_login() {
  is_a_public_ecr_registry

  if [[ $? -eq $SUCCESS ]]; then
    aws ecr-public get-login-password --region "$ECR_REGION" | helm registry login --username AWS --password-stdin $ECR_URL
  else
    aws ecr get-login-password --region "$ECR_REGION" | helm registry login --username AWS --password-stdin $ECR_URL
  fi
}

# Loads configuration and logs in to a Helm registry.
#
helm_login() {
  _load_ecr_config || die "Error while loading configuration file!"
  say "Using ECR registry url: $ECR_URL. (region: $ECR_REGION)."
  _helm_login || die "Failed to log in to the ECR registry."
}

# Generic user confirmation function
# 
confirm() {
  read -p "$@ (yes/no)" yn
  case yn in
    yes) ;;
    *)
      say "Aborting..."
      exit $FAILURE
      ;;
  esac
}
