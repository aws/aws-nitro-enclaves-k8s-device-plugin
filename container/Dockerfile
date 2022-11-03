# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

FROM golang:alpine as builder

RUN apk update

WORKDIR build_dir/
COPY *.go ./
COPY go.sum ./

RUN go mod init k8s-ne-device-plugin && \
    go mod tidy && \
    go mod vendor

RUN CGO_ENABLED=0 go build -a -ldflags='-s -w -extldflags="-static"' .

# ----------------------------------------------------

FROM scratch as device_plugin

COPY --from=builder /go/build_dir/k8s-ne-device-plugin /usr/bin/k8s-ne-device-plugin

CMD ["/usr/bin/k8s-ne-device-plugin","-logtostderr=true","-v=0"]
