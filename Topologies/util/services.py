import subprocess
import os

class ServicesUtil:

    UTIL_PREFIX = "[ServicesUtil] "

    def __init__(self):
        self.cwd = os.getcwd()
        print(self.UTIL_PREFIX+self.cwd)

    def start_Endpoint_TP(self, controllerAddr="localhost", controllerPort=4433, tpAddr="30.0.0.1", endpointTPPort=6633):
        '''
            "controllerAddr": "localhost",
            "controllerPort": "4433",
            "tpAddr": "1.2.3.4",
            "endpointTPPort": "6633"
        '''
        try:
            os.chdir(os.path.join(self.cwd, "../Endpoint_TP"))
            print(self.UTIL_PREFIX+os.getcwd())
            cmd = "./start_custom_Endpoint_TP.py"
            cmd += " --controllerAddr "+str(controllerAddr)
            cmd += " --controllerPort "+str(controllerPort)
            cmd += " --tpAddr "+str(tpAddr)
            cmd += " --endpointTPPort "+str(endpointTPPort)
            out = subprocess.check_output(cmd, universal_newlines=True, shell=True, cwd=os.path.join(self.cwd, "../Endpoint_TP"))
        except subprocess.CalledProcessError as e:
            print(self.UTIL_PREFIX+"ERROR: Failed to start Endpoint TP")
            print(e)
            exit(1)

        


#try:
#        out = subprocess.check_output("sudo ip route get "+ip_addr, universal_newlines=True, shell=True)
#        out_array = out.split(' ')
#        interface_next = False
#        for word in out_array:
#            if interface_next:
#                return word
#            if word == 'dev':
#                interface_next = True
#    except subprocess.CalledProcessError as e:
#        print("IP-Address "+ip_addr+" is invalid.")
#        print("Return the default interface")
#        return get_default_interface()