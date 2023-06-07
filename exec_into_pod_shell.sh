#!/bin/bash

kubectl exec -i -t $1 --container simple-torch-inference-bert -- /bin/bash