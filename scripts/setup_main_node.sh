#!/bin/bash

PRIVATE_IP=""

# Loop through the command-line arguments
for arg in "$@"; do
  case $arg in
    --private-ip=*)
      # Extract the private-ip value
      PRIVATE_IP="${arg#*=}"
      ;;
    *)
      # Ignore other arguments or add error handling if needed
      ;;
  esac
done

# Check if private-ip is not provided or is empty
if [[ -z $PRIVATE_IP ]]; then
  echo "ERROR: Private IP is required. Please provide the --private-ip argument."
  exit 1
fi


# Use the CIDR based on the pod network plugin you are using. This is the CIDR for Calico plugin
POD_NETWORK_CIDR=192.168.0.0/16


sudo kubeadm init --apiserver-advertise-address=$PRIVATE_IP --pod-network-cidr=$POD_NETWORK_CIDR


mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
