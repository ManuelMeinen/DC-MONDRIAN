#!/usr/bin/python3
from code_base.types import Packet
import io
from code_base.conn_state import ConnectionState
from code_base.transfer_module import TransferModule
import cProfile
import pstats
import logging
import os
import subprocess
import time

import psutil
from code_base.const import Const


class Benchmarking:
    
    LOG_PATH = "benchmarking.log"
    RES_PATH = "benchmarking_res/"
    LOG_TO_FILE = True

    def __init__(self):
        subprocess.call("rm "+self.LOG_PATH, shell=True)
        #create a logger
        self.logger = logging.getLogger('Benchmarking')
        #set logger level
        self.logger.setLevel(logging.INFO)
        handler = logging.FileHandler(self.LOG_PATH)
        # create a logging format
        formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
        handler.setFormatter(formatter)
        self.logger.addHandler(handler)

    def log(self, msg):
            if self.LOG_TO_FILE:
                self.logger.info(str(msg))
            else:
                print(str(msg))
    
    def start_mondrian_controller(self):
        '''
        Start a Mondrian Controller in the background
        '''
        self.log("Start Mondrian Controller at: https://:"+str(4433))
        try:
            cmd = "main"
            self.file = open(os.path.join(Const.BASE_PATH, "log/MONDRIAN_Controller_Port_"+str(4433)), "w")
            self.proc=subprocess.Popen(os.path.join(os.path.join(Const.BASE_PATH, "MONDRIAN_Controller"),cmd), universal_newlines=True, shell=True, stdout=self.file, stderr=self.file)
            self.log("MONDRIAN Controller process started")
            time.sleep(5)
        except subprocess.CalledProcessError as e:
            self.log("ERROR: Failed to start MONDRIAN Controller")
            print(e)
            exit(1)

    def stop_mondrian_controller(self):
        '''
        Stop the Mondrian Controller
        '''
        process = psutil.Process(self.proc.pid)
        for proc in process.children(recursive=True):
            proc.kill()
        process.kill()
        self.file.close()
        self.log("MONDRIAN Controller process terminated")
    
    def start_transfer_module(self, tpAddr="200.0.0.1", controllerAddr="localhost", controllerPort="4433"):
        '''
        Start a transfer
        '''
        self.transfer_module = TransferModule(tpAddr=tpAddr, controllerAddr=controllerAddr, controllerPort=controllerPort, refresh_interval=3000, verbose=False)
        self.log("Transfer Module started")
        # Wait for the fetcher to fetch all the data
        time.sleep(10)
    
    def start_connection_state(self):
        '''
        Start a connection state
        '''
        self.conn_state = ConnectionState()
        self.log("Connection State started")

    def write_res(self, name, s):
        '''
        Write the results in stream s to file name
        '''
        with open(name, 'w+') as f:
            f.write(s.getvalue())

    
    def bench(self, name, func, N=10000, warmup=100):
        '''
        Run function func of name name N times and benchmark it
        '''
        self.log("Run benchmarking for function "+str(name)+" for "+str(N)+" repetitions with a warmup of "+str(warmup))
        res_path = self.RES_PATH+name+".bench"
        profiler = cProfile.Profile()
        # Do some warmup
        for n in range(warmup):
            func()
        # START OF BENCHMARKED CODE
        profiler.enable()
        for n in range(N):
            func()
        profiler.disable()
        # END OF BENCHMARKED CODE
        s = io.StringIO()
        stats = pstats.Stats(profiler, stream=s).sort_stats('cumtime')
        stats.print_stats()
        self.write_res(res_path, s)
        self.log("Benchmarking done! Results can be found in "+str(res_path))
    
    def reset_subnets(self):
        '''
        set the subnets into the original state
        '''
        self.transfer_module.fetcher.refresh_subnets()
        time.sleep(5)
    
    def reset_policies(self):
        '''
        set the policies into the original state
        '''
        self.transfer_module.fetcher.refresh_policies()
        time.sleep(5)
    
    def scale_up_subnets(self, scale_factor):
        '''
        scale up the subnets by a factor of scale_factor
        '''
        self.transfer_module.fetcher.subnets = self.transfer_module.fetcher.subnets*scale_factor
    
    def scale_up_policies(self, scale_factor):
        '''
        scale up the policies by a factor of scale_factor
        '''
        self.transfer_module.fetcher.policies = self.transfer_module.fetcher.policies*scale_factor
    


if __name__=='__main__':
    bench = Benchmarking()
    bench.start_mondrian_controller()
    bench.start_transfer_module()
    bench.start_connection_state()
    try:
        packet = Packet("10.0.1.0", "20.0.2.0", destPort=100, srcPort=80, proto="UDP")
        scale_factors = [1, 2, 4, 8, 16, 32]
        N=1000000
        for f in scale_factors:
            bench.scale_up_subnets(f)
            bench.scale_up_policies(f)
            bench.log("#Subnets = "+str(len(bench.transfer_module.fetcher.subnets)))
            bench.log("#Policies = "+str(len(bench.transfer_module.fetcher.policies)))
            bench.bench("find_zone_f_"+str(f)+"_N_"+str(N), lambda: bench.transfer_module.find_zone('10.0.1.0'), N=N)
            bench.bench("check_packet_f_"+str(f)+"_N_"+str(N), lambda: bench.transfer_module.check_packet(packet), N=N)
            bench.reset_subnets()
            bench.reset_policies()
            
        
        print(len(bench.transfer_module.fetcher.subnets))
        print(len(bench.transfer_module.fetcher.policies))
        bench.reset_policies()
        bench.reset_subnets()
        print(len(bench.transfer_module.fetcher.subnets))
        print(len(bench.transfer_module.fetcher.policies))
    finally:
        bench.stop_mondrian_controller()
