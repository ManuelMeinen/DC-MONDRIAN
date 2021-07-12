#!/usr/bin/python3
import json
import sys
import subprocess
import os
from code_base.const import Const

PATH_TO_CONFIG_FILE = "Endpoint_TP/config/config.json"
BASE_PATH = Const.BASE_PATH#"/home/mmeinen/polybox/code/DC-MONDRIAN"
ENDPOINT_TP_PREFIX = "[EndpointTP] "

def update_config_file(argv):
    try:
        data = get_config()
        data = parse_cli_args(data, argv)
        with open(get_config_file_path(), "w") as jsonFile:
            json.dump(data, jsonFile, indent=4)
    except json.JSONDecodeError as e:
        print(ENDPOINT_TP_PREFIX+"ERROR: Updating config.json failed")
        exit(1)

def parse_cli_args(data, argv):
    arg_dict = {}
    for i in range(len(argv)):
        if argv[i][0:2]=='--':
            arg_dict[str(argv[i][2:])] = str(argv[i+1])
    for key, value in arg_dict.items():
        try:
                data[key]=str(value)
        except KeyError as e:
            print(ENDPOINT_TP_PREFIX+str(key)+" is not a config.json key")
    return data

def get_config():
    try:
        with open(get_config_file_path(), "r") as jsonFile:
            data = json.load(jsonFile)
            return data
    except json.JSONDecodeError as e:
        print(ENDPOINT_TP_PREFIX+"ERROR: Reading config.json file failed")
        exit(1)

def get_config_file_path():      
    path = os.path.join(BASE_PATH, PATH_TO_CONFIG_FILE)
    return path
    

if __name__=='__main__':
    update_config_file(argv=sys.argv)
    data = get_config()
    endpoint_tp = [os.path.join(os.path.join(BASE_PATH,"Endpoint_TP"),"main.py")]
    other_ryu_services = [os.path.join(BASE_PATH, "Other_RYU_Services/simple_switch_13.py")]
    config = ["--ofp-tcp-listen-port",str(data["endpointTPPort"])]
    subprocess.run(["ryu-manager"]+endpoint_tp+other_ryu_services+config)