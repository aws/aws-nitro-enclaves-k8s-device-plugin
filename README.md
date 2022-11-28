# Introduction

The Nitro Enclaves [Device Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) gives your pods and containers the ability to access the [Nitro Enclaves device driver](https://docs.kernel.org/virt/ne_overview.html). The device plugin works with both [Amazon EKS](https://aws.amazon.com/eks/) and self-managed Kubernetes nodes.

[AWS Nitro Enclaves](https://aws.amazon.com/ec2/nitro/nitro-enclaves/) is an [Amazon EC2](https://aws-content-sandbox.aka.amazon.com/ec2/) capability that enables customers to create isolated compute environments to further protect and securely process highly sensitive data within their EC2 instances.

# Prerequisites
To utilize this device plugin, you will need:

  - A configured Kubernetes cluster.
  - At least one enclave-enabled node available in the cluster. An enclave-enabled node is an EC2 instance with the **EnclaveOptions** parameter set to **true**. For more information on creating an enclaving an enclave-enabled node, review the using [Nitro Enclaves with EKS user guide](https://docs.aws.amazon.com/enclaves/latest/user/kubernetes.html).

To build the plugin, you will need:
  - Docker

# Usage
To deploy the device plugin to your Kubernetes cluster, use the following command:
```
kubectl -f apply https://raw.githubusercontent.com/aws/aws-nitro-enclaves-k8s-device-plugin/main/aws-nitro-enclaves-k8s-ds.yaml
```

After deploying the device plugin, use labelling to enable the device plugin on a particular node:
```
kubectl label node <node-name> aws-nitro-enclaves-k8s-dp=enabled
```

To see list of the nodes that have plugin enabled, use the following command:
```
kubectl get nodes --show-labels | grep aws-nitro-enclaves-k8s-dp=enabled
```

To disable the plugin on a particular node, use the following command:
```
kubectl label node <node-name> aws-nitro-enclaves-k8s-dp-
```

# Building the Device Plugin
To build the device plugin from its sources, use the following command:

```
./scripts/build.sh
````

After successfully running the script, the device plugin will be built as a Docker image with the name `aws-nitro-enclaves-k8s-device-plugin`.

# Running Nitro Enclaves in a Kubernetes Cluster

There is a guide available on how to run Nitro Enclaves in EKS clusters. See this [link](https://github.com/aws/aws-nitro-enclaves-with-k8s) to learn more.

# License
This project is licensed under the Apache-2.0 License.
