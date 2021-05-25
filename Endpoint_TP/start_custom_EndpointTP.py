import json
import sys
import subprocess

PATH_TO_CONFIG_FILE = "./config/config.json"

def update_config_file(argv):
    try:
        data = get_config()
        data = parse_cli_args(data, argv)
        with open(PATH_TO_CONFIG_FILE, "w") as jsonFile:
            json.dump(data, jsonFile, indent=4)
    except json.JSONDecodeError as e:
        print("ERROR: Updating config.json failed")
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
            print(str(key)+" is not a config.json key")
    return data

def get_config():
    try:
        with open(PATH_TO_CONFIG_FILE, "r") as jsonFile:
            data = json.load(jsonFile)
            return data
    except json.JSONDecodeError as e:
        print("ERROR: Reading config.json file failed")
        exit(1)

if __name__=='__main__':
    update_config_file(argv=sys.argv)
    data = get_config()
    endpoint_tp = ["main.py"]
    other_ryu_services = ["../Other_RYU_Services/simple_switch_13.py"]
    config = ["--ofp-tcp-listen-port",str(data["endpointTPPort"])]
    subprocess.run(["ryu-manager"]+endpoint_tp+other_ryu_services+config)