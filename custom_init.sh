#!/bin/bash

./agent --max-batchsize=128 --max-latency=100 --service-port=80 &

uvicorn app.main:app --host 0.0.0.0 --port 80
