#!/bin/bash -e

TOP_DIR=$(dirname $(realpath $0))/..
docker build --target builder -t ne-k8s-device-plugin-build:latest $TOP_DIR -f container/Dockerfile
docker build --target device_plugin -t aws-nitro-enclaves-k8s-device-plugin:latest $TOP_DIR -f container/Dockerfile