import subprocess
import os
import threading
import psutil # install with sudo python3 -m pip install psutil
import time

class ServicesUtil:

    UTIL_PREFIX = "[ServicesUtil] "
    BASE_PATH = "/home/mmeinen/polybox/code/DC-MONDRIAN" #TODO: needs to be set whenever run somewhere else... (not elegant but works...)

    def __init__(self):
        self.proc = []
        self.log = {}

    def start_Mondrian_Controller(self, controllerAddr="", controllerPort=4433):
        '''
        Start a Mondrian Controller in the background
        '''
        print(self.UTIL_PREFIX+"Start Mondrian Controller at: https://"+str(controllerAddr)+":"+str(controllerPort))
        try:
            cmd = "main"
            #cmd += " -controllerAddr "+str(controllerAddr) 
            #NOTE: We actually don't set the controller address in the MONDRIAN Controller --> like this it is reachable both via lo and docker0
            #cmd += " -controllerPort "+str(controllerPort) #TODO: why doesn't that work?...
            #print(cmd)
            if controllerPort != 4433:
                print(self.UTIL_PREFIX+"WARNING: running the MONDRIAN Controller on a differen port is currently not supported due to a bug.")
            self.log[str(controllerPort)] = open(os.path.join(self.BASE_PATH, "log/MONDRIAN_Controller_Port_"+str(controllerPort)), "w")
            self.proc.append(subprocess.Popen(os.path.join(os.path.join(self.BASE_PATH, "MONDRIAN_Controller"),cmd), universal_newlines=True, shell=True, stdout=self.log[str(controllerPort)], stderr=self.log[str(controllerPort)]))
            print(self.UTIL_PREFIX+"MONDRIAN Controller process started")
        except subprocess.CalledProcessError as e:
            print(self.UTIL_PREFIX+"ERROR: Failed to start MONDRIAN Controller")
            print(e)
            exit(1)

    def start_Endpoint_TP(self, controllerAddr="localhost", controllerPort=4433, tpAddr="30.0.0.1", endpointTPPort=6633):
        '''
        Start an Endpoint TP in the background
        '''
        print(self.UTIL_PREFIX+"Start Endpoint TP process at port "+str(endpointTPPort)+" for site "+str(tpAddr))
        try:
            cmd = "start_custom_EndpointTP.py"
            cmd += " --controllerAddr "+str(controllerAddr)
            cmd += " --controllerPort "+str(controllerPort)
            cmd += " --tpAddr "+str(tpAddr)
            cmd += " --endpointTPPort "+str(endpointTPPort)
            self.log[str(endpointTPPort)] = open(os.path.join(self.BASE_PATH, "log/Endpoint_TP_Port_"+str(endpointTPPort)), "w")
            self.proc.append(subprocess.Popen(os.path.join(os.path.join(self.BASE_PATH, "Endpoint_TP"),cmd), universal_newlines=True, shell=True, stdout=self.log[str(endpointTPPort)], stderr=self.log[str(endpointTPPort)]))
            print(self.UTIL_PREFIX+"Endpoint TP process started")
            time.sleep(1) # To prevent race condition while reading config.json file (happens rarely)
        except subprocess.CalledProcessError as e:
            print(self.UTIL_PREFIX+"ERROR: Failed to start Endpoint TP")
            print(e)
            exit(1)
        
        
    def kill_processes(self):
        '''
        Kill all Endpoint TPs, which are running in the background
        '''
        print(self.UTIL_PREFIX+"Kill all Endpoint TP processes and MONDRIAN Controllers")
        for p in self.proc:
            process = psutil.Process(p.pid)
            for proc in process.children(recursive=True):
                proc.kill()
            process.kill()
        for log in self.log:
            self.log[log].close()
            
    

