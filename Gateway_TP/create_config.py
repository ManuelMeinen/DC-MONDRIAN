#!/bin/python3
import json
import sys

if __name__=='__main__':
    hostname = str(sys.argv[1])
    tp_addr = str(sys.argv[2])
    config = {}
    config["hostname"]=hostname
    config["log_dir"]="/vol1/log/"
    config["tp_addr"]=tp_addr
    config["base_path"]=""
    with open("config/config.json", "w") as f:
        f.write(json.dumps(config, indent=4))