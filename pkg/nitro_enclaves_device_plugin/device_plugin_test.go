// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package nitro_enclaves_device_plugin

import (
	"k8s-ne-device-plugin/pkg/config"
	"testing"
)

// generateDeviceID should always generate different device
// IDs after each call. The expected result needs to
// be ≤given-device-name><nb-of-calls-made-to-the-func>
// nitro_enclaves0 ... nitro_enclaves99
func TestIncrementalDeviceIdGenerationSuccess(t *testing.T) {
	deviceName := "dummy_device"
	expected := "dummy_device_0"

	id := generateDeviceID(deviceName)

	if expected != id {
		t.Fatalf("Expected %s but got invalid id: %s!", expected, id)
		return
	}

	deviceIdCounter = 99
	_ = generateDeviceID(deviceName)
	deviceName = "nitro_enclaves"
	expected = "nitro_enclaves_100"
	id = generateDeviceID(deviceName)

	if expected != id {
		t.Fatalf("Expected %s but got invalid id: %s!", expected, id)
		return
	}
}

func TestValidateDeviceNameSuccess(t *testing.T) {
	deviceIdCounter = 50
	// per default limited to max 4 devices
	p := NewNitroEnclavesDevicePlugin(&config.PluginConfig{})

	expected := "nitro_enclaves_50"

	if expected != p.dev[0].ID {
		t.Fatalf("Expected %s but got invalid id: %s!", expected, p.dev[0].ID)
		return
	}
}

func TestValidateDeviceMaxDefaultNumber(t *testing.T) {
	// per default limited to 4 devices
	p := NewNitroEnclavesDevicePlugin(&config.PluginConfig{})

	if len(p.dev) > 4 {
		t.Fatalf("Expected 4 devices but got %d!", len(p.dev))
		return
	}
}

func TestValidateDeviceCustomNumber(t *testing.T) {
	p := NewNitroEnclavesDevicePlugin(&config.PluginConfig{MaxEnclavesPerNode: 3})

	if len(p.dev) != 3 {
		t.Fatalf("Expected 3 devices but got %d!", len(p.dev))
		return
	}
}
