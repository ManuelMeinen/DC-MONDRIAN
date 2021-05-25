import json
class Const:
    def __init__(self):
        self.init_const()

    controllerAddr = "NOT SET"
    controllerPort = "NOT SET"
    PATH_TO_CONFIG_FILE = "./config/config.json"
    tpAddr = "NOT SET"

    TCP_PROTO = "TCP"
    UDP_PROTO = "UDP"

    @classmethod
    def init_const(self):
        '''
        Initialize the constants according to the config.json file
        '''
        try:
            with open(self.PATH_TO_CONFIG_FILE, "r") as jsonFile:
                data = json.load(jsonFile)
                Const.controllerAddr = data["controllerAddr"]
                Const.controllerPort = data["controllerPort"]
                Const.tpAddr = data["tpAddr"]
        except json.JSONDecodeError as e:
            print("[Const] ERROR: Reading config.json failed!")
            exit(1)