#!/bin/bash


ab -c 1000 -n 15000 -e exp1_batch.csv -p ./request_body_example.json -T application/json http://bert.traefik.local/predict
