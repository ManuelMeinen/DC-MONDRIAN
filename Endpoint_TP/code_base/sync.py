class Synchronizer:

    '''
    This class is passed in as a context by the Endpoint TP into the RYU app_manager.
    Like this an instance of this object can be shared between the Endpoint TP and any other RYU App (e.g. simple_switch_13).
    With this, the Endpoint TP can signal to the following components if the traffic is allowed (meaning that it's ok to send 
    Packet-out messages) or if it should be droped (meaning that Packet-out messages need to be avoided by the following apps)
    '''

    def __init__(self):
        self.status = True

    def allow(self):
        '''
        Traffic is allowed according to the Endpoint TP
        '''
        self.status = True

    def drop(self):
        '''
        Traffic gets droped according to the Endpoint TP
        '''
        self.status = False
    
    def is_allowed(self):
        '''
        Is the traffic is allowed according to the Endpoint TP?
        '''
        return self.status
    
    def log_status(self, logger):
        '''
        Log the current status of the Synchronizer
        '''
        if self.status:
            logger.info("[Synchronizer] Traffic is Allowed")
        else:
            logger.info("[Synchronizer] Drop Traffic")


