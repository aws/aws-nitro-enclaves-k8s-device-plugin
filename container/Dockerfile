# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build the device plugin from sources
FROM amazonlinux:2023 as builder

RUN yum install golang-1.24.9 -y

WORKDIR build_dir/
COPY go.mod go.sum ./
COPY pkg ./pkg
COPY cmd ./cmd
# provide package cache for subsequent builds
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 go build -a -ldflags='-s -w -extldflags="-static"' -o k8s-ne-device-plugin ./cmd/k8s-device-plugin

# Create a bare minimum image that only contains the device plugin binary.
FROM scratch as device_plugin

COPY --from=builder /build_dir/k8s-ne-device-plugin /usr/bin/k8s-ne-device-plugin

CMD ["/usr/bin/k8s-ne-device-plugin", "-logtostderr=true", "-v=0"]
