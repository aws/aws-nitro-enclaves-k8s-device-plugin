// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package nitro_enclaves_device_monitor

import (
	"errors"
	"os"
	"testing"
	"time"
)

type DummyDevicePlugin struct {
	IBasicDevicePlugin
	startError error
}

func (d *DummyDevicePlugin) Start() error {
	return d.startError
}

func (d *DummyDevicePlugin) Stop() {
}

func (d *DummyDevicePlugin) ResourceName() string {
	return "aws.ec2.nitro/dummy_device"
}

func TestNoChangeOfStateAfterPluginFailsToStart(t *testing.T) {
	nepm := &NitroEnclavesPluginMonitor{
		devicePlugin: &DummyDevicePlugin{startError: errors.New("Some failure")},
	}

	nepm.setState(PluginIdle)
	run(nepm)

	if nepm.state() != PluginIdle {
		t.Fatal("Expected the state = PluginIdle, but got ", nepm.state())
		t.FailNow()
	}
}

// Whenever the Kubelet socket is recreated, the plugin needs a restart.
func TestIntegrationValidatePluginNeedsARestart(t *testing.T) {
	dp := "/tmp/"
	ksn := dp + "dummy.domain.socket"

	nepm := &NitroEnclavesPluginMonitor{
		devicePlugin:      &DummyDevicePlugin{},
		devicePluginPath:  dp,
		kubeletSocketName: ksn,
	}

	// Remove the dummy socket file if exists
	os.Remove(ksn)
	result := nepm.Init()

	if result != nil {
		t.Fatal("Error while initializing plugin monitor.")
		t.FailNow()
	}

	nepm.setState(PluginRunning)
	go run(nepm)
	// Reschedule
	time.Sleep(100 * time.Millisecond)

	// Create a dummy socket file
	fdesc, _ := os.Create(ksn)

	err := fdesc.Close()
	if err != nil {
		t.Logf("Error while creating dummy socket file.")
	}
	defer os.Remove(ksn)

	// Wait for the monitor state to change.
	for i := 0; i < 10; i++ {
		if nepm.state() == PluginRestarting {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	// Test requires direct access to /tmp folder.
	// Might fail on macOS natively - run in container instead.
	if nepm.state() != PluginRestarting {
		t.Fatal("Socket file is generated, but the plugin didn't restart!")
		t.FailNow()
	}
}
