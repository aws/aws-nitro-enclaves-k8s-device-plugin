module k8s-ne-device-plugin

go 1.26.0

toolchain go1.26.7

require (
	github.com/fsnotify/fsnotify v1.10.1
	github.com/golang/glog v1.2.5
	golang.org/x/net v0.59.0
	google.golang.org/grpc v1.83.2
	k8s.io/kubelet v0.37.0
)

require (
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260911204522-f61a6ca850bd // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
