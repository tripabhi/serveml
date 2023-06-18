#!/bin/bash

./agent --max-batchsize=128 --max-latency=100 --service-port=3000 &

python app/main.py
