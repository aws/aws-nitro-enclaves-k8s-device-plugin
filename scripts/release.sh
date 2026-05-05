#!/bin/bash
# Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
set -e

for arg in "$@"; do
  case $arg in
    --non-interactive)
      export NE_NON_INTERACTIVE=true
      ;;
    *)
      echo "[ERROR] Unknown flag: $arg" >&2
      echo "Usage: $0 [--non-interactive]" >&2
      exit 1
      ;;
  esac
done

source "$(dirname $(realpath $0))/common.sh"
current_folder="$(dirname $(realpath $0))"

# BuildKit flags propagated to build_docker.sh.
#   --load     makes the built image available in the local image store so
#              push_docker.sh can tag and push it. Override with BUILDX_LOAD=""
#              for engines that store images locally by default (e.g. Podman).
#   BUILDX_CACHE is optional (e.g. --cache-from type=gha --cache-to type=gha,mode=max).
export BUILDX_LOAD="${BUILDX_LOAD:---load}"
export BUILDX_CACHE="${BUILDX_CACHE:-}"

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
