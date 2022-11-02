// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	glog.V(0).Info("Loading K8s Nitro Enclaves device plugin...")

	devicePlugin := NewNitroEnclavesDevicePlugin()
	monitor := NewNitroEnclavesMonitor(devicePlugin)

	if monitor == nil {
		glog.Error("Error while initializing NE plugin monitor!")
		os.Exit(1)
	}
	
	monitor.Run()
}
