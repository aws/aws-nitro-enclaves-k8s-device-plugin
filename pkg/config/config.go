// Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"github.com/golang/glog"
	"os"
	"strconv"
)

type PluginConfig struct {
	MaxEnclavesPerNode      int
	EnclaveCPUAdvertisement bool
}

const (
	// EC2 instance with nitro_option enabled, can support upto 4 enclaves.
	// https://docs.aws.amazon.com/enclaves/latest/user/multiple-enclaves.html
	maxEnclavesPerInstance = 4
)

func (c *PluginConfig) Validate() error {
	if c.MaxEnclavesPerNode <= 0 || c.MaxEnclavesPerNode > maxEnclavesPerInstance {
		c.MaxEnclavesPerNode = maxEnclavesPerInstance
		return fmt.Errorf("max devices per node must be greater than 0 and smaller or equal to %v - set value to max", maxEnclavesPerInstance)
	}
	return nil
}
func LoadConfig() *PluginConfig {
	// config parameters are primarily sourced via environment variables
	config := &PluginConfig{}

	var enclaveCPUAdvertisement bool
	enclaveCPUAdvertisement, err := strconv.ParseBool(os.Getenv("ENCLAVE_CPU_ADVERTISEMENT"))
	if err != nil {
		glog.Errorf("error parsing ENCLAVE_CPU_ADVERTISEMENT: %v", err)
		glog.Infof("setting ENCLAVE_CPU_ADVERTISEMENT to: %v", false)
		enclaveCPUAdvertisement = false
	}
	config.EnclaveCPUAdvertisement = enclaveCPUAdvertisement

	maxDevices, err := strconv.Atoi(os.Getenv("MAX_ENCLAVES_PER_NODE"))
	if err != nil {
		glog.Errorf("error parsing MAX_DEVICES_PER_NODE: %v", err)
		glog.Infof("Setting MAX_DEVICES_PER_NODE to: %v", maxEnclavesPerInstance)
		maxDevices = maxEnclavesPerInstance
	}
	config.MaxEnclavesPerNode = maxDevices

	return config
}
