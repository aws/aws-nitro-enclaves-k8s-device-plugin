# Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

FROM scratch

COPY k8s-ne-device-plugin /usr/bin/k8s-ne-device-plugin

CMD ["/usr/bin/k8s-ne-device-plugin","-logtostderr=true","-v=5"]
