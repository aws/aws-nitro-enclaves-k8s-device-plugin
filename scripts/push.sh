#!/bin/bash
# Temporary file for development purposes only. Will be removed.

set -eu pipefail

aws ecr get-login-password --region eu-central-1 | docker login --username AWS --password-stdin 709843417989.dkr.ecr.eu-central-1.amazonaws.com
$(dirname $0)/build.sh
sync
docker image rm -f plugin_devel
docker build -t plugin_devel:latest .
docker tag plugin_devel:latest 709843417989.dkr.ecr.eu-central-1.amazonaws.com/plugin_devel:latest
docker push 709843417989.dkr.ecr.eu-central-1.amazonaws.com/plugin_devel:latest
