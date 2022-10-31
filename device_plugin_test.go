// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
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
	p := NewNitroEnclavesDevicePlugin()

	expected := "nitro_enclaves_50"

	if expected != p.dev[0].ID {
		t.Fatalf("Expected %s but got invalid id: %s!", expected, p.dev[0].ID)
		return
	}
}
