import time
import sys
sys.path.append("..") #TODO figure out wtf is wrong with python imports
from code_base.const import Const
class Stats:
    '''
    This class is used to keep track of the number of packet-in messages per delta_t
    '''
    def __init__(self, hard_timeout, idle_timeout, delta_t=60):
        self.hard_timeout = hard_timeout
        self.idle_timeout = idle_timeout
        self.time = time.time()
        self.count = 0
        self.delta_t = delta_t
        self.From = 0
        self.To = delta_t
        self.res_path = Const.BASE_PATH+"/Endpoint_TP/benchmarking_res/packet-in_report_"+str(Const.endpointTPPort)+"_HARD_TIMEOUT_"+str(self.hard_timeout)+"_IDLE_TIMEOUT_"+str(self.idle_timeout)+".bench"
        with open(self.res_path, 'w+') as f:
            f.write("")
        self.write("FROM    TO  Packets\n")
        self.write("-------------------\n")

    def tick(self):
        '''
        Is invoked whenever the event we want to observe occured
        '''
        while time.time()-self.time > self.delta_t:
            self.write(str(self.From)+"s     "+str(self.To)+"s     "+str(self.count)+"\n")
            self.count = 0
            self.From = self.To
            self.To = self.To+self.delta_t
            self.time = self.time+self.delta_t
        self.count += 1
    
    def write(self, msg):
        '''
        Write the results in stream s to file name
        '''
        with open(self.res_path, 'a') as f:
            f.write(str(msg))

    