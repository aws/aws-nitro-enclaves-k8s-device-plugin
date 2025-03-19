# aws-nitro-enclaves-k8s-device-plugin

The Nitro Enclaves [Device Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) gives your pods and containers the ability to access the [Nitro Enclaves device driver](https://docs.kernel.org/virt/ne_overview.html).\
The device plugin works with both [Amazon EKS](https://aws.amazon.com/eks/) and self-managed Kubernetes nodes.\
[AWS Nitro Enclaves](https://aws.amazon.com/ec2/nitro/nitro-enclaves/) is an [Amazon EC2](https://aws-content-sandbox.aka.amazon.com/ec2/) capability that enables customers to create isolated compute environments to further protect and securely process highly sensitive data within their EC2 instances.

-------
## Prerequisites

To utilize this device plugin, you will need:

- A configured Kubernetes cluster.
- At least one enclave-enabled node available in the cluster. An enclave-enabled node is an EC2 instance with the **EnclaveOptions** parameter set to **true**.\
For more information on creating an enclave-enabled EKS worker node, review the using [Nitro Enclaves with EKS user guide](https://docs.aws.amazon.com/enclaves/latest/user/kubernetes.html).

To build the plugin, you will need:
- Docker

----
## Configuring the AWS Nitro Enclave device plugin for Kubernetes

The device plugin supports the following two configuration options via the device-plugins daemon-set environment
variables:

### MAX_ENCLAVES_PER_NODE
AWS EKS nodes can support up to `4` enclaves per node. Number can be reduced if required and evaluated by the Kubernetes scheduler.

```yaml
- name: MAX_ENCLAVES_PER_NODE
  value: "4"
```
If deployed, EKS worker nodes do advertise the number of available `enclave` resources in the following way:
```yaml
Capacity:
  aws.ec2.nitro/nitro_enclaves: 4
```

```yaml
resources:
  limits:
    aws.ec2.nitro/nitro_enclaves: "1"
  requests:
    aws.ec2.nitro/nitro_enclaves: "1"
```



### ENCLAVE_CPU_ADVERTISEMENT
Advertise the number of `offline` CPUs on a specific EKS worker node. The number of offline CPUs reflect the number of CPUs allocated by the Nitro allocation service during EKS worker node startup.\
By advertising the number of available CPUs, workloads can request specific amount of CPUs for their enclaves and the Kubernetes scheduler can place workloads according to available CPUs on EKS worker nodes. Set to `false` per default.

```yaml
- name: ENCLAVE_CPU_ADVERTISEMENT
  value: "false"
```

If enabled, EKS worker nodes do advertise the allocatable CPUs in the following way:
```yaml
Capacity:
  aws.ec2.nitro/nitro_enclaves_cpus: 12
```

Kubernetes workloads can request CPUs for their enclaves (e.g. 2) by adding `aws.ec2.nitro/nitro_enclaves_cpus: "2"` to the resources `limits` and `requests` sections as shown below:

```yaml
resources:
  limits:
    aws.ec2.nitro/nitro_enclaves_cpus: "2"
  requests:
    aws.ec2.nitro/nitro_enclaves_cpus: "2"
```

### Example Deployment Specification
The following snippet represents a fully populated `resources` section for a Kubernetes pod requesting access to a single enclave that requires `2Gi` of memory and access to `2` CPUs.\
Refer to the [official Using Nitro Enclaves with Amazon EKS documentation](https://docs.aws.amazon.com/enclaves/latest/user/kubernetes.html) for more information on the different options in the deployment spec.

```yaml
resources:
  limits:
    aws.ec2.nitro/nitro_enclaves: "1"
    aws.ec2.nitro/nitro_enclaves_cpus: "2"
    hugepages-1Gi: 2Gi
    cpu: 250m
  requests:
    aws.ec2.nitro/nitro_enclaves: "1"
    aws.ec2.nitro/nitro_enclaves_cpus: "2"
    hugepages-1Gi: 2Gi
```

---------

## Usage 
### Deploy Kubernetes Manifest

To deploy the device plugin to your Kubernetes cluster, use the following command:

```shell
kubectl apply -f https://raw.githubusercontent.com/aws/aws-nitro-enclaves-k8s-device-plugin/main/aws-nitro-enclaves-k8s-ds.yaml
```

After deploying the device plugin, use labelling to enable the device plugin on a particular node:

```shell
kubectl label node <node-name> aws-nitro-enclaves-k8s-dp=enabled
```

To see list of the nodes that have plugin enabled, use the following command:

```shell
kubectl get nodes --show-labels | grep aws-nitro-enclaves-k8s-dp=enabled
```

To disable the plugin on a particular node, use the following command:

```shell
kubectl label node <node-name> aws-nitro-enclaves-k8s-dp-
```

### Deployment via Helm Chart

To deploy the Helm chart for the device plugin to your Kubernetes cluster refer
to [Helm Readme](./helm/README.md)


---------
## Building the Device Plugin Locally

To build the device plugin from its sources, use the following command:

```shell
./scripts/build.sh
````

After successfully running the script, the device plugin will be built as a Docker image with the name `aws-nitro-enclaves-k8s-device-plugin`.

---------

## Running Nitro Enclaves in a Kubernetes Cluster

There is a guide available on how to run Nitro Enclaves in EKS clusters. See
this [link](https://github.com/aws/aws-nitro-enclaves-with-k8s) to learn more.

# License

This project is licensed under the Apache-2.0 License.
