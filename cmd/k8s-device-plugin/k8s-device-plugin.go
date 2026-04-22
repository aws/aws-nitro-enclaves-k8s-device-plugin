// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"k8s-ne-device-plugin/pkg/config"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_cpu_plugin"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_device_monitor"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_device_plugin"
	"os"
)

// These variables are populated at build time via -ldflags -X.
// Defaults apply when building without version injection (e.g. plain `go build`).
var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("k8s-ne-device-plugin version %s (built: %s)\n", version, buildDate)
		os.Exit(0)
	}

	glog.V(0).Infof("Starting K8s Nitro Enclaves device plugin %s (built: %s)", version, buildDate)

	// load config from manifest file and validate
	pluginConfig := config.LoadConfig()

	// create nitro enclave device, pass it to monitor and start in background
	enclaveDevicePlugin := nitro_enclaves_device_plugin.NewNitroEnclavesDevicePlugin(pluginConfig)
	enclaveDeviceMonitor := nitro_enclaves_device_monitor.NewNitroEnclavesMonitor(enclaveDevicePlugin)
	if enclaveDeviceMonitor == nil {
		glog.Error("Error while initializing Nitro Enclave Device plugin monitor!")
		os.Exit(1)
	}

	// create and start nitro enclave cpu device in background to advertise available cpus
	if pluginConfig.EnclaveCPUAdvertisement {
		cpuDevicePlugin := nitro_enclaves_cpu_plugin.NewNitroEnclavesCPUDevicePlugin(pluginConfig)
		cpuDeviceMonitor := nitro_enclaves_device_monitor.NewNitroEnclavesMonitor(cpuDevicePlugin)
		if cpuDeviceMonitor == nil {
			glog.Error("Error while initializing Nitro Enclave CPU Device plugin monitor!")
			os.Exit(1)
		}
		go cpuDeviceMonitor.Run()
	}

	// start nitro enclave device plugin and start monitoring loop, main thread is active as long as enclave device
	// plugin is running and healthy, otherwise terminate and have k8s restart container
	enclaveDeviceMonitor.Run()
}
