#from Endpoint_TP.code_base.transfer_module import TransferModule
import os
import time
from ryu.base import app_manager
from ryu.controller import ofp_event
from ryu.controller.handler import CONFIG_DISPATCHER, MAIN_DISPATCHER
from ryu.controller.handler import set_ev_cls
from ryu.ofproto import ofproto_v1_0, ofproto_v1_2, ofproto_v1_3, ofproto_v1_4, ofproto_v1_5

from ryu.lib.packet import in_proto
from ryu.lib.packet import ipv4
from ryu.lib.packet import tcp
from ryu.lib.packet import udp

from ryu.base import app_manager
from ryu.controller import ofp_event
from ryu.controller.handler import CONFIG_DISPATCHER, MAIN_DISPATCHER
from ryu.controller.handler import set_ev_cls
from ryu.ofproto import ofproto_v1_3
from ryu.lib.packet import packet
from ryu.lib.packet import ethernet
from ryu.lib.packet import ether_types
from ryu.app.simple_switch_stp_13 import SimpleSwitch13

from code_base.types import Packet, proto_dict, Policy, Zone, Subnet
from code_base.const import Const
from code_base.sync import Synchronizer
from code_base.conn_state import ConnectionState
from code_base.transfer_module import ESTABLISHED_RESPONSE, TransferModule, ESTABLISHED, FORWARDING, DROP, INTRA_ZONE, DEFAULT
from code_base.stats import Stats

class EndpointTP(app_manager.RyuApp):

    OFP_VERSIONS = [ofproto_v1_3.OFP_VERSION]
    TABLE_ID = 0
    # Timeouts are in seconds and 0 menas it never times out
    #IDLE_TIMEOUT = 60*60
    #HARD_TIMEOUT = 10*60
    BENCHMARKING = True #Change to False if stats should be turned off

    def log(self, msg):
        if self.verbose:
            self.logger.info(Const.ENDPOINT_TP_PREFIX+str(msg))

    _CONTEXTS = {
    'synchronizer': Synchronizer
    }

    def __init__(self, *args, **kwargs):
        super(EndpointTP, self).__init__(*args, **kwargs)
        self.synchronizer = kwargs['synchronizer']
        c = Const(self.logger)
        self.verbose = True
        self.module = TransferModule(tpAddr=Const.tpAddr, controllerAddr=Const.controllerAddr, 
                                    controllerPort=Const.controllerPort, logger=None, verbose=self.verbose)
        self.conn_state = ConnectionState(logger=self.logger, verbose=self.verbose)
        if self.BENCHMARKING:      
            self.stats = Stats(hard_timeout=Const.HARD_TIMEOUT, idle_timeout=Const.IDLE_TIMEOUT, delta_t=1)
            
        
    
    @set_ev_cls(ofp_event.EventOFPSwitchFeatures, CONFIG_DISPATCHER)
    def switch_features_handler(self, ev):
        datapath = ev.msg.datapath
        ofproto = datapath.ofproto
        parser = datapath.ofproto_parser
        # install table-miss flow entry
        #
        # We specify NO BUFFER to max_len of the output action due to
        # OVS bug. At this moment, if we specify a lesser number, e.g.,
        # 128, OVS will send Packet-In with invalid buffer_id and
        # truncated packet data. In that case, we cannot output packets
        # correctly.  The bug has been fixed in OVS v2.1.0.
        match = parser.OFPMatch()
        actions = [parser.OFPActionOutput(ofproto.OFPP_CONTROLLER,
                                          ofproto.OFPCML_NO_BUFFER)]
        self.add_flow(datapath, 0, match, actions) # NOTE: This flow is not allowed to time out --> hard_timeout = idle_timeout = 0
        

    @set_ev_cls(ofp_event.EventOFPPacketIn, MAIN_DISPATCHER)
    def packet_in_handler(self, ev): 
        msg = ev.msg
        datapath = msg.datapath
        ofproto = datapath.ofproto
        parser = datapath.ofproto_parser
        in_port = msg.match['in_port']

        pkt = packet.Packet(msg.data)
        eth = pkt.get_protocols(ethernet.ethernet)[0]

        dstip = None
        srcip = None
        dstport = None
        srcport = None
        proto = None
        self.synchronizer.allow()
        if eth.ethertype == ether_types.ETH_TYPE_IP:
            self.log(30*'-'+" New Packet-In "+30*'-')
            ip = pkt.get_protocol(ipv4.ipv4)
            if ip == None:
                self.log("WARNING: Non IPv4 Packet detected. IPv6 is currently not supported.")
                self.synchronizer.log_status(self.logger)
                return
            srcip = ip.src
            dstip = ip.dst
            proto = ip.proto
            l3_proto = None
            
            #  if TCP Protocol
            if proto == in_proto.IPPROTO_TCP:
                t = pkt.get_protocol(tcp.tcp)
                dstport = t.dst_port
                srcport = t.src_port
                l3_proto = Const.TCP_PROTO        
            #  If UDP Protocol 
            elif proto == in_proto.IPPROTO_UDP:
                u = pkt.get_protocol(udp.udp)
                dstport = u.dst_port
                srcport = u.src_port
                l3_proto = Const.UDP_PROTO
            else:
                # Directly store the proto number in the packet
                l3_proto = proto
            
            packet_in = Packet(destIP=dstip, srcIP=srcip, destPort=dstport, srcPort=srcport, proto=l3_proto)
            src_net, dest_net, packet_in, action, established_resp = self.module.check_packet(packet=packet_in)
            if src_net == None or dest_net == None:
                self.log("src_net or dest_net not found in the MONDRIAN Controller --> Packet can't be handled")
                self.synchronizer.log_status(self.logger)
                self.synchronizer.allow()
                return
            if self.BENCHMARKING:
                self.stats.tick()
            match_dict = self.createMatchDict(in_port=in_port, src_net=src_net, dest_net=dest_net, packet_in=packet_in)
            match = parser.OFPMatch(**match_dict)
            # actions = [ESTABLISHED, ESTABLISHED_RESPONSE, FORWARDING, DROP, INTRA_ZONE, DEFAULT]
            if established_resp !=None: 
                if self.conn_state.check_with_state(init_net=dest_net, resp_net=src_net, init_port=packet_in.destPort, 
                                                    resp_port=packet_in.srcPort, proto=packet_in.proto):
                    # This packet belongs to an established connection
                    actions = []
                    instructions = [parser.OFPInstructionGotoTable(table_id = self.TABLE_ID+1)]
                    self.log("Packet classification: "+str(established_resp)+" --> GOTO next table")
                    self.log(packet_in.to_string())
                else:
                    # Connection not in state --> contact the controller for following packets to not miss a connection setup
                    self.synchronizer.drop()
                    return
            else:    
                # From here on it's for sure not an established response
                if action == DROP or action == DEFAULT:
                    # Drop the traffic (which is the default)
                    actions = []
                    instructions = []
                    self.log("Packet classification: "+str(action)+" --> DROP")
                    self.log(packet_in.to_string())
                elif action == ESTABLISHED:
                    # Let the VNF in the next flow table handle the traffic
                    actions = []
                    instructions = [parser.OFPInstructionGotoTable(table_id = self.TABLE_ID+1)]
                    self.log("Packet classification: "+str(action)+" --> GOTO next table")
                    self.log(packet_in.to_string())
                    self.conn_state.add_to_state(init_net=src_net, resp_net=dest_net, init_port=packet_in.srcPort, resp_port=packet_in.destPort, proto=packet_in.proto)
                else:
                     # Let the VNF in the next flow table handle the traffic
                    actions = []
                    instructions = [parser.OFPInstructionGotoTable(table_id = self.TABLE_ID+1)]
                    self.log("Packet classification: "+str(action)+" --> GOTO next table")
                    self.log(packet_in.to_string()) 

            if msg.buffer_id != ofproto.OFP_NO_BUFFER:
                self.add_flow(datapath, 1, match, actions, msg.buffer_id, instructions=instructions, idle_timeout=Const.IDLE_TIMEOUT, hard_timeout=Const.HARD_TIMEOUT)
            else:
                self.add_flow(datapath, 1, match, actions, instructions=instructions, idle_timeout=Const.IDLE_TIMEOUT, hard_timeout=Const.HARD_TIMEOUT)
            if action == DROP or action == DEFAULT:
                self.synchronizer.drop()
            else:
                self.synchronizer.allow()
            
  

    def add_flow(self, datapath, priority, match, actions, buffer_id=None, instructions=[], idle_timeout=0, hard_timeout=0):
        ofproto = datapath.ofproto
        parser = datapath.ofproto_parser
    
        inst = [parser.OFPInstructionActions(ofproto.OFPIT_APPLY_ACTIONS,
                                             actions)]+instructions
        
        if buffer_id:
            mod = parser.OFPFlowMod(datapath=datapath, buffer_id=buffer_id,
                                    priority=priority, match=match,
                                    instructions=inst, table_id=self.TABLE_ID, 
                                    idle_timeout=idle_timeout, hard_timeout=hard_timeout) 
        else:
            mod = parser.OFPFlowMod(datapath=datapath, priority=priority,
                                    match=match, instructions=inst, table_id=self.TABLE_ID, 
                                    idle_timeout=idle_timeout, hard_timeout=hard_timeout)
        datapath.send_msg(mod)
        
    

    def createMatchDict(self, in_port, src_net, dest_net, packet_in):
        # Basic src and dest match
        match_dict = {
                    'in_port':in_port,
                    'eth_type':ether_types.ETH_TYPE_IP,
                    'ipv4_src':src_net,
                    'ipv4_dst':dest_net
                }
        # Handle TCP packet
        if packet_in.proto==Const.TCP_PROTO:
            match_dict['ip_proto'] = in_proto.IPPROTO_TCP
            if packet_in.srcPort != None:
                match_dict['tcp_src'] = packet_in.srcPort
            if packet_in.destPort != None:
                match_dict['tcp_dst'] = packet_in.destPort 
        # Handle UDP packet
        elif packet_in.proto==Const.UDP_PROTO:
            match_dict['ip_proto'] = in_proto.IPPROTO_UDP
            if packet_in.srcPort != None:
                match_dict['udp_src'] = packet_in.srcPort
            if packet_in.destPort != None:
                match_dict['udp_dst'] = packet_in.destPort 
        else:
        # Handle any other packet type (i.e. ICMP)
            match_dict['ip_proto'] = packet_in.proto
        return match_dict
        
           
        

    