# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build the device plugin from sources
FROM amazonlinux:2023 as builder

RUN dnf install golang-1.25.8 -y

WORKDIR build_dir/
COPY go.mod go.sum ./
COPY pkg ./pkg
COPY cmd ./cmd
COPY RELEASE ./
# provide package cache for subsequent builds
RUN go mod download

RUN VERSION=$(tr -d '[:space:]' < RELEASE) && \
    BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
    CGO_ENABLED=0 go build -a \
      -ldflags="-s -w -extldflags=-static \
        -X main.version=${VERSION} \
        -X main.buildDate=${BUILD_DATE}" \
      -o k8s-ne-device-plugin ./cmd/k8s-device-plugin

# Create a bare minimum image that only contains the device plugin binary.
FROM scratch as device_plugin

COPY --from=builder /build_dir/k8s-ne-device-plugin /usr/bin/k8s-ne-device-plugin

CMD ["/usr/bin/k8s-ne-device-plugin", "-logtostderr=true", "-v=0"]
