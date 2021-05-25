import requests
import sys
import threading
import time
sys.path.append("..") #TODO figure out wtf is wrong with python imports
from code_base.types import Packet, proto_dict, Policy, Zone, Subnet
from code_base.const import Const #controllerAddr, controllerPort


subnets_url = "https://"+Const.controllerAddr+":"+Const.controllerPort+"/api/get-subnets"
transitions_url = "https://"+Const.controllerAddr+":"+Const.controllerPort+"/api/get-transitions"


class Fetcher:
    '''
            This Fetcher starts a deamon that fetches the relevant subnets and policies periodically.
            In order to avoid race conditions use the get_subnets and get_policies functions
    '''
    def __init__(self, tpAddr, refresh_interval=30):
        self.tpAddr = tpAddr
        self.refresh_interval = refresh_interval
        self.subnets = []
        self.policies = []
        self.subnet_lock = threading.Lock()
        self.policy_lock = threading.Lock()
        self.refresh_subnets()
        self.refresh_policies()
        self.deamon = threading.Thread(target=self.refresh_deamon)
        self.deamon.daemon = True
        self.deamon.start()

    
    def refresh_deamon(self):
        '''
        Run the fetching deamon
        '''
        print("[Fetcher] Refresh deamon started")
        while True:
            time.sleep(self.refresh_interval)
            self.refresh_subnets()
            self.refresh_policies()    
            

    def get_subnets(self):
        '''
        Returns the subnets as they are stored in the fetcher
        '''
        self.subnet_lock.acquire()
        try:
            print("[Fetcher] subnet lock acquired")
            subnets = self.subnets
        finally:
            self.subnet_lock.release()
            print("[Fetcher] subnet lock released")
            return subnets
    
    def get_policies(self):
        '''
        Returns the policies as they are stored in the fetcher
        '''
        self.policy_lock.acquire()
        try:
            print("[Fetcher] policy lock acquired")
            policies = self.policies
        finally:
            self.policy_lock.release()
            print("[Fetcher] policy lock released")
            return policies

    def refresh_subnets(self):
        '''
        Lock the subnets and refresh them
        '''
        self.subnet_lock.acquire()
        try:
            print("[Fetcher] subnet lock acquired")
            fresh_subnets = self.fetch_subnets() 
            if fresh_subnets != None:
                self.subnets = fresh_subnets
            else:
                print("[Fetcher] WARNING! No new subnets fetched. Local version is preserved and might be out of date.")
        finally:
            self.subnet_lock.release()
            print("[Fetcher] subnet lock released")
    
    def refresh_policies(self):
        '''
        Lock the policies and refresh them
        '''
        self.policy_lock.acquire()
        try:
            print("[Fetcher] policy lock acquired")
            fresh_policies = self.fetch_policies()
            if fresh_policies != None:
                self.policies = fresh_policies
            else:
                print("[Fetcher] WARNING! No new policies fetched. Local version is preserved and might be out of date.")
        finally:
            self.policy_lock.release()
            print("[Fetcher] policy lock released")

    def fetch_subnets(self):
        '''
        Fetch the latest subnets from the controller
        '''
        # Proceed, only if no error:
        resp_dict = {}
        try:
            resp = requests.post(subnets_url, data=self.tpAddr, verify=False) #TODO figure out how to do that safely
            resp.raise_for_status()
            # Decode JSON response into a Python dict:
            resp_dict = resp.json()
        except requests.exceptions.HTTPError as e:
            print("Bad HTTP status code:", e)
            return None
        except requests.exceptions.RequestException as e:
            print("Network error:", e)
            return None
        print("[Fetcher] Subnets fetched successfully")
        subnets = []
        for line in resp_dict:
            netAddr = line['CIDR']
            zoneID = int(line['ZoneID'])
            tpAddr = line['TPAddr']
            net = Subnet(netAddr=netAddr, zoneID=zoneID, tpAddr=tpAddr)
            subnets.append(net)
        return subnets
    
    def fetch_policies(self):
        '''
        Fetch the latest policies from the controller
        '''
        # Proceed, only if no error:
        resp_dict = {}
        try:
            resp = requests.post(transitions_url, data=self.tpAddr, verify=False) #TODO figure out how to do that safely
            resp.raise_for_status()
            # Decode JSON response into a Python dict:
            resp_dict = resp.json()
        except requests.exceptions.HTTPError as e:
            print("Bad HTTP status code:", e)
            return None
        except requests.exceptions.RequestException as e:
            print("Network error:", e)
            return None
        print("[Fetcher] Policies fetched successfully")
        transitions = []
        for line in resp_dict:
            policyID = int(line['PolicyID'])
            action = line['Action']
            destZoneID=None
            srcZoneID=None
            destPort=None
            srcPort=None
            proto=None
            if int(line['Dest']) != 0:
                destZoneID = int(line['Dest'])
            if int(line['Src']) != 0:
                srcZoneID = int(line['Src'])
            if int(line['DestPort']) != 0:
                destPort = int(line['DestPort'])
            if int(line['SrcPort']) != 0:
                srcPort = int(line['SrcPort'])
            if line['Proto'] != '':
                proto = line['Proto']
            t = Policy(policyID=policyID, action=action, destZoneID=destZoneID, srcZoneID=srcZoneID, destPort=destPort, srcPort=srcPort, proto=proto)
            transitions.append(t)
        return transitions
