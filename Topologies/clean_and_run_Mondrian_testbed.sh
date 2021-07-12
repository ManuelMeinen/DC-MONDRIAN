#!/bin/bash
docker ps -q --filter name=mn.* | xargs docker stop
docker ps -a -q --filter name=mn.* | xargs docker rm

./Mondrian_testbed.py