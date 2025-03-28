// Copyright 2022 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nitro_enclaves_device_plugin

import (
	"errors"
	"google.golang.org/grpc/credentials/insecure"
	"k8s-ne-device-plugin/pkg/config"
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
	// check if socketPath does exist and delete otherwise do nothing
	_, err := os.Stat(nedp.pdef.socketPath())
	if err == nil {
		err = os.Remove(nedp.pdef.socketPath())
		if err != nil {
			glog.Errorf("Error removing socket file: %s", err)
		}
	}
}

// Register the device plugin with Kubelet.
func (nedp *NitroEnclavesDevicePlugin) register(kubeletEndpoint, resourceName string) error {
	glog.V(0).Info("Attempting to connect to kubelet...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//lint:ignore SA1019 grpc.WithBlock is deprecated, not supported by grpc.NewClient
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.DialUnix("unix", nil, &net.UnixAddr{Name: addr, Net: "unix"})
		}),
	}

	//lint:ignore SA1019 grpc.DialContext is deprecated // todo replace by grpc.NewClient
	conn, err := grpc.DialContext(
		ctx,
		kubeletEndpoint,
		opts...,
	)
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			glog.Errorf("Error closing connection to kubelet: %s", err)
		}
	}(conn)

	if err != nil {
		glog.Errorf("Couldn't connect to kubelet! (Reason: %s)", err)
		return err
	}

	glog.V(0).Info("Connected to kubelet")

	client := pluginapi.NewRegistrationClient(conn)
	_, err = client.Register(ctx, &pluginapi.RegisterRequest{
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

	return errors.New("gRPC server initialization timed out")
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
func (nedp *NitroEnclavesDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
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
	err := s.Send(&pluginapi.ListAndWatchResponse{Devices: nedp.dev})
	if err != nil {
		return err
	}

	//TODO: Device health check goes here
	<-nedp.stop
	return nil

}

// PreStartContainer is called, if indicated by Device Plugin during registration phase,
// before each container start. Device plugin can run device specific operations
// such as resetting the device before making devices available to the container.
func (nedp *NitroEnclavesDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
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
	go func() {
		err := nedp.server.Serve(sock)
		if err != nil {
			if nedp.stop != nil {
				glog.Errorf("Error while serving device plugin: %v", err)
				close(nedp.stop)
			}
		}
	}()
	err = nedp.waitForServerReady(devicePluginServerReadyTimeout)
	if err != nil {
		return err
	}

	if err = nedp.register(pluginapi.KubeletSocket, nedp.pdef.resourceName()); err != nil {
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
func NewNitroEnclavesDevicePlugin(config *config.PluginConfig) *NitroEnclavesDevicePlugin {

	if err := config.Validate(); err != nil {
		glog.Errorf("invalid plugin config: %v", err)
	}

	glog.V(0).Infof("Initializing Nitro Enclaves device plugin with following params: %v", config)

	// devs slice, determines the pluginapi.ListAndWatchResponse, which lets the kubelet know about the available/allocatable "aws.ec2.nitro/nitro_enclaves" devices
	// in a k8s worker node. Number of devices, in this context does not represent number of "nitro_enclaves" device files present in the host,
	// instead it can be interpreted as number pods that can share the same host device file. The same host device file "nitro_enclaves",
	// can be mounted into multiple pods, which can be used to run an enclave.
	// This lets us schedule 2 or more pods requiring nitro_enclaves device on the same k8s node/EC2 instance.
	devs := []*pluginapi.Device{}
	for i := 0; i < config.MaxEnclavesPerNode; i++ {
		devs = append(devs, &pluginapi.Device{
			ID:     generateDeviceID(deviceName),
			Health: pluginapi.Healthy,
		})
	}
	glog.V(0).Infof("Enclave devices added: %v", config.MaxEnclavesPerNode)

	return &NitroEnclavesDevicePlugin{
		dev:    devs,
		pdef:   &NEPluginDefinitions{},
		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}
