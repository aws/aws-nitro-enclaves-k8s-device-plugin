// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"github.com/golang/glog"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_device_monitor"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_device_plugin"
	"os"
)

func main() {
	flag.Parse()
	glog.V(0).Info("Loading K8s Nitro Enclaves device plugin...")


	// create nitro enclave device, pass it to monitor and start in background
	devicePlugin := nitro_enclaves_device_plugin.NewNitroEnclavesDevicePlugin()
	monitor := nitro_enclaves_device_monitor.NewNitroEnclavesMonitor(devicePlugin)
	if monitor == nil {
		glog.Error("Error while initializing NE plugin monitor!")
		os.Exit(1)
	}

	monitor.Run()
}
