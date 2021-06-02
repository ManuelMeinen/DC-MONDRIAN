import logging
import sys
sys.path.append("..") #TODO figure out wtf is wrong with python imports
from code_base.fetcher import Fetcher
from code_base.types import Packet
from code_base.const import Const
from ipaddress import ip_network, ip_address

ESTABLISHED = "established"
FORWARDING = "forwarding"
DROP = "drop"
INTRA_ZONE = "intra_zone"
DEFAULT = "default"
ACTIONS = [ESTABLISHED, FORWARDING, DROP, INTRA_ZONE, DEFAULT]

class TransferModule:

    def __init__(self, tpAddr, controllerAddr, controllerPort, logger=None):
        self.tpAddr = tpAddr
        self.logger = logger
        self.fetcher = Fetcher(tpAddr=tpAddr, controllerAddr=controllerAddr, controllerPort=controllerPort, logger=logger)
    

    def check_packet(self, packet):
        '''
        Checks a packet for a matching policy. It returns the match with the highest priority and breaks ties with the policyID.
        The match can be reconstructed using the src and dest subnet address, the packet 
        (for protocol and port) and an ACTION (drop, forwarding, established, intra_zone, default)
        Priority: srcZone +5, destZone +5, srcPort +1, destPort +2, proto +1
        return src_net, dest_net, packet, ACTION
        '''
        policies = self.fetcher.get_policies()
        src, src_net = self.find_zone(packet.srcIP)
        dest, dest_net = self.find_zone(packet.destIP)
        if src == dest:
            # Same Zone traffic
            return src_net, dest_net, packet, INTRA_ZONE
        # Iterate over all policies and check for the match with the highest policyID
        matching_policy = None
        highest_priority = 0
        for policy in policies:
            priority = 0
            policy.print_policy()
            match=False
            if policy.srcZoneID == None or policy.srcZoneID == src:
                match = True
                priority += 5
            else:
                match = False
            if match==True and (policy.destZoneID == None or policy.destZoneID == dest):
                match = True
                priority += 5
            else: 
                match=False
            if match==True and (policy.srcPort == None or policy.srcPort == packet.srcPort):
                match = True
                priority += 1
            else: 
                match=False
            if match==True and (policy.destPort == None or policy.destPort == packet.destPort):
                match = True
                priority += 2
            else: 
                match=False
            if match==True and (policy.proto == None or policy.proto == packet.proto):
                match = True
                priority += 1
            else: 
                match=False
            if match:
                policy.print_policy()
                packet.print_packet()
                if matching_policy == None or highest_priority <= priority:
                    if highest_priority == priority:
                        # Break ties
                        if matching_policy.policyID < policy.policyID:
                            matching_policy = policy
                            highest_priority = priority
                    else:
                        matching_policy = policy
                        highest_priority = priority
                        continue
            # check for established rule
            priority = 0
            if policy.action == "established":
                match = True
            else: 
                match=False
            if match==True and (policy.srcZoneID == None or policy.srcZoneID == dest):
                match = True
                priority += 5
            else: 
                match=False
            if match==True and (policy.destZoneID == None or policy.destZoneID == src):
                match = True
                priority += 5
            else: 
                match=False
            if match==True and (policy.srcPort == None or policy.srcPort == packet.destPort):
                match = True
                priority += 1
            else: 
                match=False
            if match==True and (policy.destPort == None or policy.destPort == packet.srcPort):
                match = True
                priority += 2
            else: 
                match=False
            if match==True and (policy.proto == None or policy.proto == packet.proto):
                match = True
                priority += 1
            else: 
                match=False
            if match:
                policy.print_policy()
                packet.print_packet()
                if matching_policy == None or highest_priority <= priority:
                    if highest_priority == priority:
                        # Break ties
                        if matching_policy.policyID < policy.policyID:
                            matching_policy = policy
                            highest_priority = priority
                    else:
                        matching_policy = policy
                        highest_priority = priority
                        continue
            # Depending on the policy found (if any) return the right info
            if matching_policy == None:
                # No matching policy --> default
                return src_net, dest_net, packet, DEFAULT
            else:
                # There is a match --> drop, forwarding, established
                return src_net, dest_net, packet, matching_policy.action
                
            


    def find_zone(self, ip_addr):
        '''
        Return the zoneID of the zone in which ip_addr is 
        WARNING: This function returns None if the zone wasn't found
        '''
        subnets = self.fetcher.get_subnets()
        longest_prefix = 0
        matching_zone = None
        matching_subnet = None
        for subnet in subnets:
            net = ip_network(subnet.netAddr)
            if ip_address(ip_addr) in net:
                if matching_zone == None:
                    matching_zone = subnet.zoneID
                    longest_prefix = net.prefixlen
                    matching_subnet = subnet.netAddr
                else: 
                    if longest_prefix<net.prefixlen:
                        matching_zone = subnet.zoneID
                        longest_prefix = net.prefixlen
                        matching_subnet = subnet.netAddr
        if matching_zone == None:
            self.logger.info(Const.ENDPOINT_TP_PREFIX+"[Transfer Module] ERROR: Zone not found for "+str(ip_addr))
        return matching_zone, matching_subnet

