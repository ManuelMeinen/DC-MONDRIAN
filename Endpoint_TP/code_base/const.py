import json
class Const:
    def __init__(self):
        pass

    controllerAddr = "localhost"
    controllerPort = "4433"
    PATH_TO_CONFIG_FILE = "./config/config.json"
    tpAddr = "1.2.3.4"

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
                self.controllerAddr = data["controllerAddr"]
                self.controllerPort = data["controllerPort"]
                self.tpAddr = data["tpAddr"]
        except json.JSONDecodeError as e:
            print("[Const] ERROR: Reading config.json failed!")
            exit(1)