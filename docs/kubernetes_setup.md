# Kubernetes Setup

## Prerequisites
For the current installation, you require 3 Ubuntu 22.04 machines, with different hostnames.
You can change the hostnames by running:
``` bash
sudo hostnamectl set-hostname $NEW_HOST_NAME
```
Make sure they have the latest OS updates, by running:
``` bash
sudo apt update && sudo apt upgrade -y
```

## Install Dependencies
Our Kubernetes installation requires the following core dependencies:
- `containerd`
- `kubectl`
- `kubeadm`
- `kubelet`

Run the following on all 3 nodes.
``` bash
./scripts/install_dependencies.sh
```

## Setup Control Plane
One of the nodes in the cluster will contain the control plane components like the Kube API server. Run the following from the node that will act as the control-plane node:
``` bash
./scripts/setup_main_node.sh --private-ip=$PRIVATE_IP
```
Here, use the private IP of your machine as the `$PRIVATE_IP`

## Add Pod Network Plugin
We use Calico as the pod networking add-on. To install calico, run the following:
``` bash
./scripts/install_calico.sh
```

## Join Cluster
Run the following from the master node or control-plane node:
``` bash
kubeadm token create --print-join-command
```
Copy the output and run it from the worker nodes.

## Verify the cluster
Run the following to see if the nodes are ready in your cluster:
``` bash
kubectl get nodes
```
