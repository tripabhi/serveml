#!/bin/bash


ab -c 1000 -n 1000 -p ./request_body_example.json -T application/json http://10.96.195.40/predict
