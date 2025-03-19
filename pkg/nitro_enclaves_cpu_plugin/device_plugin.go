// Copyright 2025 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nitro_enclaves_cpu_plugin

import (
	"errors"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"k8s-ne-device-plugin/pkg/config"
	"net"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	deviceName                     = "nitro_enclaves_cpus"
	devicePluginServerReadyTimeout = 10
	deviceOfflineCPUsPath          = "/sys/devices/system/cpu/offline"
)

var cpuIdCounter = 0

type IBasicDevicePlugin interface {
	Start() error
	Stop()
	socketPath() string
	resourceName() string
}

// NitroEnclavesCPUDevicePlugin implements the Kubernetes device plugin API
type NitroEnclavesCPUDevicePlugin struct {
	devices []*pluginapi.Device

	stop chan interface{}

	server *grpc.Server
	pluginapi.DevicePluginServer
	IBasicDevicePlugin
}

func (necdp *NitroEnclavesCPUDevicePlugin) socketPath() string {
	return pluginapi.DevicePluginPath + deviceName + ".sock"
}

func (necdp *NitroEnclavesCPUDevicePlugin) ResourceName() string {
	return "aws.ec2.nitro/" + deviceName
}

// determineAdvisableCPUs reads the number of offline cpus from /sys/devices/system/cpu/offline
func determineAdvisableCPUs(data string) (int, error) {

	// Handle empty/unknown case
	content := strings.TrimSpace(data)
	if content == "" {
		return 0, nil
	}

	total := 0
	ranges := strings.Split(content, ",")

	for _, r := range ranges {
		parts := strings.Split(r, "-")
		switch len(parts) {
		case 1:
			// Single CPU
			// ensure that parts is a valid number
			_, err := strconv.Atoi(parts[0])
			if err != nil {
				return 0, fmt.Errorf("invalid CPU number: %s, parsing caused error: %w", r, err)
			}
			total++
		case 2:
			// CPU range
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				return 0, fmt.Errorf("invalid CPU range: %s", r)
			}
			total += end - start + 1
		default:
			return 0, fmt.Errorf("malformed CPU range: %s", r)
		}
	}

	return total, nil
}

func generateEnclaveCPUID(deviceName string) string {
	ctr := cpuIdCounter
	cpuIdCounter++
	return deviceName + "_" + strconv.Itoa(ctr)
}

// Register the device plugin with Kubelet.
func (necdp *NitroEnclavesCPUDevicePlugin) register(kubeletEndpoint, resourceName string) error {
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
			glog.Errorf("Error closing connection to kubelet: %v", err)
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
		Endpoint:     path.Base(necdp.socketPath()),
		ResourceName: resourceName,
	})

	return err
}

// Ensure that the gRPC server of the device plugin is ready to serve.
func (necdp *NitroEnclavesCPUDevicePlugin) waitForServerReady(timeout int) error {
	for i := 0; i < timeout; i++ {
		info := necdp.server.GetServiceInfo()
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
func (necdp *NitroEnclavesCPUDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}

	for _, req := range reqs.ContainerRequests {
		responses.ContainerResponses = append(responses.ContainerResponses, &pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"NITRO_ENCLAVES_CPUS": strconv.Itoa(len(req.DevicesIDs)),
			},
		})
	}

	return &responses, nil
}

// GetDevicePluginOptions returns options to be communicated with Device Manager.
func (*NitroEnclavesCPUDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// GetPreferredAllocation returns a preferred set of devices to allocate
// from a list of available ones. The resulting preferred allocation is not
// guaranteed to be the allocation ultimately performed by the
// devicemanager. It is only designed to help the devicemanager make a more
// informed allocation decision when possible.
func (*NitroEnclavesCPUDevicePlugin) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// ListAndWatch returns a stream of List of Devices
// Whenever a Device state change or a Device disappears, ListAndWatch
// returns the new list
func (necdp *NitroEnclavesCPUDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	err := s.Send(&pluginapi.ListAndWatchResponse{Devices: necdp.devices})
	if err != nil {
		return err
	}

	<-necdp.stop
	return nil

}

// PreStartContainer is called, if indicated by Device Plugin during registration phase,
// before each container start. Device plugin can run device specific operations
// such as resetting the device before making devices available to the container.
func (necdp *NitroEnclavesCPUDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (necdp *NitroEnclavesCPUDevicePlugin) releaseResources() {
	necdp.server = nil
	// check if socketPath does exist and delete otherwise do nothing
	_, err := os.Stat(necdp.socketPath())
	if err == nil {
		err = os.Remove(necdp.socketPath())
		if err != nil {
			glog.Errorf("Error removing socket file: %s", err)
		}
	}
}

// Start device plugin server
func (necdp *NitroEnclavesCPUDevicePlugin) Start() error {
	necdp.releaseResources()
	glog.V(0).Info("Starting Nitro Enclaves CPU device plugin server...")

	sock, err := net.Listen("unix", necdp.socketPath())
	if err != nil {
		glog.Error("Error while creating socket: ", necdp.socketPath())
		return err
	}

	necdp.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(necdp.server, necdp)
	go func() {
		err := necdp.server.Serve(sock)
		if err != nil {
			if necdp.stop != nil {
				glog.Errorf("Error while serving device plugin: %v", err)
				close(necdp.stop)
			}
		}
	}()
	err = necdp.waitForServerReady(devicePluginServerReadyTimeout)
	if err != nil {
		return err
	}

	if err = necdp.register(pluginapi.KubeletSocket, necdp.resourceName()); err != nil {
		glog.Errorf("Error while registering cpu device plugin with kubelet! (Reason: %s)", err)
		necdp.Stop()
		return err
	}
	glog.V(0).Info("Registered cpu device plugin with Kubelet: ", necdp.resourceName())

	return nil
}

// Stop device plugin server
func (necdp *NitroEnclavesCPUDevicePlugin) Stop() {
	close(necdp.stop)
	if necdp.server != nil {
		necdp.server.Stop()
		necdp.releaseResources()
		necdp.server = nil
	}
	glog.V(0).Infof("CPU device plugin stopped. (Socket: %s)", necdp.socketPath())
}

// NewNitroEnclavesCPUDevicePlugin returns an initialized NitroEnclavesCPUDevicePlugin
func NewNitroEnclavesCPUDevicePlugin(config *config.PluginConfig) *NitroEnclavesCPUDevicePlugin {

	if err := config.Validate(); err != nil {
		glog.Errorf("invalid CPU plugin config: %v", err)
	}

	glog.V(0).Infof("Initializing Nitro Enclaves CPU device plugin with following params: %v", config)

	// create a virtual device for each 'offline' cpu on the kubernetes worker. An offline CPU can be considered a
	// CPU that is not in use by the host OS and has thus been allocated by the AWS Nitro Enclave allocation service.
	var devs []*pluginapi.Device
	if config.EnclaveCPUAdvertisement {

		data, err := os.ReadFile(deviceOfflineCPUsPath)
		if err != nil {
			glog.V(0).Infof("Error reading offline CPU file: %v", err)
			// if error was thrown in read CPU file step, set data to empty string to have
			// determineAdvisableCPUs set availableCPUsOnInstance to 0
			data = []byte("")
		}

		availableCPUsOnInstance, err := determineAdvisableCPUs(string(data))
		if err != nil {
			glog.V(0).Infof("Error while determining advisable CPUs on the instance: %v", err)
			availableCPUsOnInstance = 0
		}

		for i := 0; i < availableCPUsOnInstance; i++ {
			devs = append(devs, &pluginapi.Device{
				ID:     generateEnclaveCPUID(deviceName),
				Health: pluginapi.Healthy,
			})
		}
		glog.V(0).Infof("Reserved CPUs for encalves added: %v", availableCPUsOnInstance)
	}

	return &NitroEnclavesCPUDevicePlugin{
		devices: devs,
		stop:    make(chan interface{}),
	}
}

func (necdp *NitroEnclavesCPUDevicePlugin) Serve() error {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Start the gRPC server
	if err := necdp.Start(); err != nil {
		return fmt.Errorf("failed to start device plugin: %v", err)
	}

	// watch for socket deletion or signals
	socketPath := necdp.socketPath()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			necdp.Stop()
			return nil
		case <-ticker.C:
			if _, err := os.Stat(socketPath); err != nil {
				glog.Info("Socket file missing, restarting plugin")
				necdp.Stop()
				if err := necdp.Start(); err != nil {
					glog.Error(err)
				}
			}
		}
	}
}
