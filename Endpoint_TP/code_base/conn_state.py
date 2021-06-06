import sys
sys.path.append("..") #TODO figure out wtf is wrong with python imports
from code_base.const import Const
class ConnectionState:

    def __init__(self, state_size = 100, logger=None, verbose=True):
        self.state_size = state_size
        self.logger = logger
        self.verbose = verbose
        self.allowed_resp = []
    
    def log(self, msg):
      if self.verbose:
          if self.logger == None:
              print(str(msg))
          else:
              self.logger.info(Const.CONNECTION_STATE+str(msg))

    def add_to_state(self, init_net, resp_net, init_port, resp_port, proto):
        '''
        Add a connection to the state
        '''
        self.allowed_resp.append((init_net, resp_net, init_port, resp_port, proto))
        self.log("Conn added to state: "+str((init_net, resp_net, init_port, resp_port, proto)))
        while len(self.allowed_resp)>self.state_size:
            # Prevent the state to become too large --> prevent DDoS attacks (Note that the state should be constant and quite small anyways)
            self.allowed_resp = self.allowed_resp[1:]
    
    def check_with_state(self, init_net, resp_net, init_port, resp_port, proto):
        '''
        Check if the connection is in the state and if yes remove it
        '''
        conn =(init_net, resp_net, init_port, resp_port, proto)
        if conn in self.allowed_resp:
            self.log("Conn in state: "+str(conn))
            self.allowed_resp.remove(conn)
            return True
        else:
            self.log("Conn is not in state! "+str(conn))
            return False
