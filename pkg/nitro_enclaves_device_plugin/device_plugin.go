// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nitro_enclaves_device_plugin

import (
	"errors"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	deviceName                     = "nitro_enclaves"
	devicePluginServerReadyTimeout = 10
	// EC2 instance with nitro_option enabled, can support upto 4 enclaves.
	// https://docs.aws.amazon.com/enclaves/latest/user/multiple-enclaves.html
	enclavesPerInstance = 4
)

var deviceIdCounter = 0

type IPluginDefinitions interface {
	socketPath() string
	devicePath() string
	resourceName() string
}

type NEPluginDefinitions struct {
	IPluginDefinitions
}

func (n *NEPluginDefinitions) socketPath() string {
	return pluginapi.DevicePluginPath + deviceName + ".sock"
}

func (n *NEPluginDefinitions) devicePath() string {
	return "/dev/" + deviceName
}

func (n *NEPluginDefinitions) resourceName() string {
	return "aws.ec2.nitro/" + deviceName
}

type IBasicDevicePlugin interface {
	Start() error
	Stop()
}

// NitroEnclavesDevicePlugin implements the Kubernetes device plugin API
type NitroEnclavesDevicePlugin struct {
	dev  []*pluginapi.Device
	pdef IPluginDefinitions

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server

	pluginapi.DevicePluginServer
	IBasicDevicePlugin
}

func generateDeviceID(deviceName string) string {
	ctr := deviceIdCounter
	deviceIdCounter++
	return deviceName + "_" + strconv.Itoa(ctr)
}
func (nedp *NitroEnclavesDevicePlugin) releaseResources() {
	nedp.server = nil
	os.Remove(nedp.pdef.socketPath())
}

// Register the device plugin with Kubelet.
func (nedp *NitroEnclavesDevicePlugin) register(kubeletEndpoint, resourceName string) error {
	glog.V(0).Info("Attempting to connect to kubelet...")

	conn, err := grpc.Dial(kubeletEndpoint, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		glog.Errorf("Couldn't connect to kubelet! (Reason: %s)", err)
		return err
	}

	glog.V(0).Info("Connected to kubelet.")

	defer conn.Close()
	client := pluginapi.NewRegistrationClient(conn)
	_, err = client.Register(context.Background(), &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(nedp.pdef.socketPath()),
		ResourceName: resourceName,
	})

	return err
}

// Ensure that the gRPC server of the device plugin is ready to serve.
func (nedp *NitroEnclavesDevicePlugin) waitForServerReady(timeout int) error {
	for i := 0; i < timeout; i++ {
		info := nedp.server.GetServiceInfo()
		if len(info) >= 1 {
			return nil
		}
		time.Sleep(time.Second)
	}

	return errors.New("gRPC server initialization timed out!")
}

// Allocate is called during container creation so that the Device
// Plugin can run device specific operations and instruct Kubelet
// of the steps to make the Device available in the container
func (nedp *NitroEnclavesDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{
			Devices: []*pluginapi.DeviceSpec{
				{
					ContainerPath: nedp.pdef.devicePath(),
					HostPath:      nedp.pdef.devicePath(),
					Permissions:   "rw",
				},
			},
		}

		for _, id := range req.DevicesIDs {
			glog.V(1).Info("Allocation request for device ID: ", id)
		}

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

// GetDevicePluginOptions returns options to be communicated with Device Manager.
func (*NitroEnclavesDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// GetPreferredAllocation returns a preferred set of devices to allocate
// from a list of available ones. The resulting preferred allocation is not
// guaranteed to be the allocation ultimately performed by the
// devicemanager. It is only designed to help the devicemanager make a more
// informed allocation decision when possible.
func (*NitroEnclavesDevicePlugin) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// ListAndWatch returns a stream of List of Devices
// Whenever a Device state change or a Device disappears, ListAndWatch
// returns the new list
func (nedp *NitroEnclavesDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: nedp.dev})

	//TODO: Device health check goes here
	<-nedp.stop
	return nil

}

// PreStartContainer is called, if indicated by Device Plugin during registeration phase,
// before each container start. Device plugin can run device specific operations
// such as resetting the device before making devices available to the container.
func (m *NitroEnclavesDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

// Start device plugin server
func (nedp *NitroEnclavesDevicePlugin) Start() error {
	nedp.releaseResources()
	glog.V(0).Info("Starting Nitro Enclaves device plugin server...")

	sock, err := net.Listen("unix", nedp.pdef.socketPath())

	if err != nil {
		glog.Error("Error while creating socket: ", nedp.pdef.socketPath())
		return err
	}

	nedp.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(nedp.server, nedp)
	go nedp.server.Serve(sock)
	nedp.waitForServerReady(devicePluginServerReadyTimeout)

	if err := nedp.register(pluginapi.KubeletSocket, nedp.pdef.resourceName()); err != nil {
		glog.Errorf("Error while registering device plugin with kubelet! (Reason: %s)", err)
		nedp.Stop()
		return err
	}

	glog.V(0).Info("Registered device plugin with Kubelet: ", nedp.pdef.resourceName())
	return nil
}

// Stop device plugin server
func (nedp *NitroEnclavesDevicePlugin) Stop() {
	if nedp.server != nil {
		nedp.server.Stop()
		nedp.releaseResources()
		glog.V(0).Infof("Device plugin stopped. (Socket: %s)", nedp.pdef.socketPath())
	}
}

// NewNitroEnclavesDevicePlugin returns an initialized NitroEnclavesDevicePlugin
func NewNitroEnclavesDevicePlugin() *NitroEnclavesDevicePlugin {
	// devs slice, determines the pluginapi.ListAndWatchResponse, which lets the kublet know about the available/allocatable "aws.ec2.nitro/nitro_enclaves" devices
	// in a k8s worker node. Number of devices, in this context does not represent number of "nitro_enclaves" device files present in the host,
	// instead it can be interpreted as number pods that can share the same host device file. The same host device file "nitro_enclaves",
	// can be mounted into multiple pods, which can be used to run an enclave.
	// This lets us to schedule 2 or more pods requiring nitro_enclaves device on the same k8s node/EC2 instance.
	devs := []*pluginapi.Device{}
	for i := 0; i < enclavesPerInstance; i++ {
		devs = append(devs, &pluginapi.Device{
			ID:     generateDeviceID(deviceName),
			Health: pluginapi.Healthy,
		})
	}
	return &NitroEnclavesDevicePlugin{
		dev:    devs,
		pdef:   &NEPluginDefinitions{},
		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}
