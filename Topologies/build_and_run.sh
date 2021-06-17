#!/bin/bash
docker ps -q --filter name=mn.* | xargs docker stop
docker ps -a -q --filter name=mn.* | xargs docker rm

cd ../Gateway_TP
./build_image.sh
cd ../Topologies
./GatewayTP_testbed.py