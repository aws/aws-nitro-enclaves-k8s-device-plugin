// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nitro_enclaves_device_monitor

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type PluginState int

const (
	PluginIdle       PluginState = 0
	PluginRunning    PluginState = 1
	PluginRestarting PluginState = 2

	pluginStartRetryTimeout = 3 * time.Second
)

type IPluginState interface {
	state() PluginState
	setState(PluginState)
}

type NitroEnclavesPluginMonitor struct {
	pluginState       PluginState
	devicePlugin      IBasicDevicePlugin
	fsWatcher         *fsnotify.Watcher
	sigWatcher        chan os.Signal
	devicePluginPath  string
	kubeletSocketName string
	IPluginState
}

func (ps PluginState) String() string {
	switch ps {
	case PluginIdle:
		return "Idle"
	case PluginRestarting:
		return "Restarting"
	case PluginRunning:
		return "Running"
	default:
		return "Unknown"
	}
}

func (nepm *NitroEnclavesPluginMonitor) state() PluginState {
	return nepm.pluginState
}

func (nepm *NitroEnclavesPluginMonitor) setState(newState PluginState) {
	nepm.pluginState = newState
}

func (nepm *NitroEnclavesPluginMonitor) Init() error {
	glog.V(0).Infof("Creating plugin monitor for %v", nepm.devicePlugin.ResourceName())
	nepm.setState(PluginIdle)

	var err error

	if nepm.fsWatcher, err = fsnotify.NewWatcher(); err != nil {
		glog.Error("Error while creating file system watcher!")
		return err
	}

	if err = nepm.fsWatcher.Add(nepm.devicePluginPath); err != nil {
		glog.Errorf("Error while accessing: %s", pluginapi.DevicePluginPath)
		defer nepm.fsWatcher.Close()
		return err
	}

	nepm.sigWatcher = make(chan os.Signal, 1)
	signal.Notify(nepm.sigWatcher, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	glog.V(0).Info("Plugin monitor has been successfully created.")

	return nil
}

func run(nepm *NitroEnclavesPluginMonitor) bool {
	cont := true

	if nepm.state() != PluginRunning {
		if err := nepm.devicePlugin.Start(); err != nil {
			// Sleep and try again as long as the monitor is running.
			time.Sleep(pluginStartRetryTimeout)
			return cont
		}
	}

	nepm.setState(PluginRunning)
	glog.V(0).Infof("%v plugin state is: %v.", nepm.devicePlugin.ResourceName(), nepm.state())

L:
	select {
	case fsEvent := <-nepm.fsWatcher.Events:
		//glog.V(0).Info("FS EVENT: ", fsEvent)
		if fsEvent.Name == nepm.kubeletSocketName {
			if fsEvent.Op&fsnotify.Create == fsnotify.Create {
				glog.V(0).Infof("Kubelet sock has been re/created. The plugin needs a restart.")
				nepm.setState(PluginRestarting)
				break L
			}
		}

	case sig := <-nepm.sigWatcher:
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			glog.V(0).Infof("Terminating plugin monitor... (Reason: \"%v\")", sig)
			nepm.devicePlugin.Stop()
			cont = false
			break L
		}
	}

	return cont
}

func (nepm *NitroEnclavesPluginMonitor) Run() {
	defer nepm.fsWatcher.Close()

	for ever := true; ever; {
		ever = run(nepm)
	}
}

type IBasicDevicePlugin interface {
	Start() error
	Stop()
	ResourceName() string
}

// Create a new plugin monitor.
func NewNitroEnclavesMonitor(nedp IBasicDevicePlugin) *NitroEnclavesPluginMonitor {
	nepm := &NitroEnclavesPluginMonitor{
		devicePlugin:      nedp,
		devicePluginPath:  pluginapi.DevicePluginPath,
		kubeletSocketName: pluginapi.KubeletSocket,
	}

	if nepm.Init() != nil {
		nepm = nil
	}

	return nepm
}
