#!/bin/bash
# Temporary file for development purposes only. Will be removed.

set -eu pipefail

URI=709843417989.dkr.ecr.eu-central-1.amazonaws.com
REPOSITORY=709843417989.dkr.ecr.eu-central-1.amazonaws.com/plugin_devel:latest

aws ecr get-login-password --region eu-central-1 | docker login --username AWS --password-stdin $URI
docker tag aws-nitro-enclaves-k8s-device-plugin:latest $REPOSITORY
docker push $REPOSITORY
