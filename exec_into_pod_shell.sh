#!/bin/bash

microk8s kubectl exec -i -t $1 --container simple-torch-inference-bert -n sml -- /bin/bash