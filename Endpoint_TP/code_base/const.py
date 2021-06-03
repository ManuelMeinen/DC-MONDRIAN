import json
import os
class Const:
    def __init__(self, logger = None):
        self.init_const()
        self.logger = logger

    controllerAddr = "NOT SET"
    controllerPort = "NOT SET"
    BASE_PATH = "/home/mmeinen/polybox/code/DC-MONDRIAN" #TODO: needs to be set whenever run somewhere else... (not elegant but works...)
    PATH_TO_CONFIG_FILE = "Endpoint_TP/config/config.json"
    
    tpAddr = "NOT SET"

    TCP_PROTO = "TCP"
    UDP_PROTO = "UDP"

    ENDPOINT_TP_PREFIX = "[EndpointTP] "
    TRANSFER_MODULE_PREFIX = ENDPOINT_TP_PREFIX+"[Transfer Module] "
    FETCHER_PREFIX = ENDPOINT_TP_PREFIX+"[Fetcher] "

    @classmethod
    def init_const(self):
        '''
        Initialize the constants according to the config.json file
        '''
        Const.PATH_TO_CONFIG_FILE = os.path.join(Const.BASE_PATH, Const.PATH_TO_CONFIG_FILE)
        try:
            with open(Const.PATH_TO_CONFIG_FILE, "r") as jsonFile:
                data = json.load(jsonFile)
                Const.controllerAddr = data["controllerAddr"]
                Const.controllerPort = data["controllerPort"]
                Const.tpAddr = data["tpAddr"]
        except json.JSONDecodeError as e:
            self.logger.info("[Const] ERROR: Reading config.json failed!")
            exit(1)