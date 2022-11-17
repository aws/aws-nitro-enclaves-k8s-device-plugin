# Introduction

This device plugin enables the use of [Nitro Enclaves](https://aws.amazon.com/ec2/nitro/nitro-enclaves/)
in your [EKS](https://aws.amazon.com/eks/) cluster.

# Prerequisites
To utilize this device plugin, you will need:

  - A configured [Kubernetes](https://kubernetes.io/) cluster on [EKS](https://aws.amazon.com/eks/)
  - An Enclave enabled [EC2](https://aws.amazon.com/ec2/features/) worker node
  - Docker

# Usage
You can install the device plugin to your **Kubernetes** cluster via the command below:
```
kubectl -f apply https://raw.githubusercontent.com/aws/aws-nitro-enclaves-k8s-device-plugin/main/aws-nitro-enclaves-k8s-ds.yaml
```

After deploying the device plugin, use labelling to enable device plugin on a particular node:
```
kubectl label node <node-name> aws-nitro-enclaves-k8s-dp=enabled
```

To see list of the nodes that have plugin enabled, use:
```
kubectl get nodes --show-labels | grep aws-nitro-enclaves-k8s-dp=enabled
```

Disabling the plugin on a particular node is possible with the command-line below:
```
kubectl label node <node-name> aws-nitro-enclaves-k8s-dp-
```

# Running Nitro Enclaves in an EKS Cluster

There is a hands-on guide available on how to run Nitro Enclaves in EKS clusters. Please check this [link](https://github.com/aws/aws-nitro-enclaves-with-k8s) to
learn more.

# License
This project is licensed under the Apache-2.0 License.
