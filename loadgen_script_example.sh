#!/bin/bash


ab -c 1000 -n 15000 -p ./request_body_example.json -T application/json http://10.152.183.147/predict
