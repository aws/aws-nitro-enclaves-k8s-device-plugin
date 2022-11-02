#!/bin/bash

ADDITIONAL_ARGS=""

if [[ "$1" = "yes" ]]; then
# Rebuild the project
ADDITIONAL_ARGS="-a"
fi 

cd $(dirname $0)/..
go mod init k8s-ne-device-plugin && go mod tidy && go mod vendor
CGO_ENABLED=0 go build ${ADDITIONAL_ARGS} -ldflags='-s -w -extldflags="-static"' .