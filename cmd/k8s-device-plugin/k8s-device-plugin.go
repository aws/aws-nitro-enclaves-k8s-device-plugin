// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"github.com/golang/glog"
	"k8s-ne-device-plugin/pkg/config"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_cpu_plugin"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_device_monitor"
	"k8s-ne-device-plugin/pkg/nitro_enclaves_device_plugin"
	"os"
)

func main() {
	flag.Parse()
	glog.V(0).Info("Loading K8s Nitro Enclaves device plugin...")

	// load config from manifest file and validate
	pluginConfig := config.LoadConfig()

	// create nitro enclave device, pass it to monitor and start in background
	devicePlugin := nitro_enclaves_device_plugin.NewNitroEnclavesDevicePlugin(pluginConfig)
	monitor := nitro_enclaves_device_monitor.NewNitroEnclavesMonitor(devicePlugin)
	if monitor == nil {
		glog.Error("Error while initializing NE plugin monitor!")
		os.Exit(1)
	}

	// create and start nitro enclave cpu device in background to advertise available cpus
	if pluginConfig.EnclaveCPUAdvertisement {
		go func() {
			devicePluginCPU := nitro_enclaves_cpu_plugin.NewNitroEnclavesCPUDevicePlugin(pluginConfig)
			if err := devicePluginCPU.Serve(); err != nil {
				glog.Errorf("Error running the NE CPU plugin: %v", err)
			}
		}()
	}

	// start nitro enclave device plugin and start monitoring loop, main thread is active as long as enclave device
	// plugin is running and healthy, otherwise terminate and have k8s restart container
	monitor.Run()
}
