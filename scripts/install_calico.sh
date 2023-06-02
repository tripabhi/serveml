#!/bin/bash

mkdir -p ../calico

# Create Tigera Operator
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.0/manifests/tigera-operator.yaml

# Download calico manifests
curl https://raw.githubusercontent.com/projectcalico/calico/v3.26.0/manifests/custom-resources.yaml -o ../calico/custom-resources.yaml

# Create Kubernetes resources for calico
kubectl create -f ../calico/custom-resources.yaml
