#from Endpoint_TP.code_base.transfer_module import TransferModule
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

from code_base.types import Packet, proto_dict, Policy, Zone, Subnet
from code_base.const import TCP_PROTO, UDP_PROTO, tpAddr
from code_base.transfer_module import TransferModule

class EndpointTP(app_manager.RyuApp):
    OFP_VERSIONS = [ofproto_v1_3.OFP_VERSION]

    def debug(self, string):
        print("*** DEBUG INFO *** "+str(string))

    def __init__(self, *args, **kwargs):
        super(EndpointTP, self).__init__(*args, **kwargs)
        self.module = TransferModule(tpAddr=tpAddr)

    @set_ev_cls(ofp_event.EventOFPPacketIn, MAIN_DISPATCHER)
    def packet_in_handler(self, ev):
        
        msg = ev.msg
        datapath = msg.datapath
        ofproto = datapath.ofproto
        parser = datapath.ofproto_parser
        in_port = msg.match['in_port']

        pkt = packet.Packet(msg.data)
        eth = pkt.get_protocols(ethernet.ethernet)[0]


        #############################
        #       Start of my stuff   #
        #############################
        dstip = None
        srcip = None
        dstport = None
        srcport = None
        proto = None

        if eth.ethertype == ether_types.ETH_TYPE_IP:
            ip = pkt.get_protocol(ipv4.ipv4)
            if ip == None:
                print("[Endpoint TP] WARNING: Non IPv4 Packet detected. IPv6 is currently not supported.")
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
                l3_proto = TCP_PROTO        
            #  If UDP Protocol 
            elif proto == in_proto.IPPROTO_UDP:
                u = pkt.get_protocol(udp.udp)
                dstport = u.dst_port
                srcport = u.src_port
                l3_proto = UDP_PROTO
            
            packet_in = Packet(destIP=dstip, srcIP=srcip, destPort=dstport, srcPort=srcport, proto=l3_proto)
            '''
            ----------- PACKET -----------
            Dest IP: 10.0.0.1
            Src IP: 10.0.0.2
            Dest Port: 9999
            Src Port: 48214
            Proto: TCP
            ------------------------------
            '''
            src_net = "10.0.0.0/8"
            dest_net = "10.0.0.0/8"
            action = "intra_zone"
            #src_net, dest_net, packet_in, action = self.module.check_packet(packet=packet_in)
            print(dest_net)
            print(src_net)
            packet_in.print_packet()
            print(action)
            #############################
            #       End of my stuff     #
            #############################

            # TODO: figure out how to adapt it to our needs and delete the unnecessary stuff... but be careful since I don't know where I fucked up last time...

            # install a flow to avoid packet_in next time
            #if out_port != ofproto.OFPP_FLOOD:
            actions = [] # DROP
            match = parser.OFPMatch(in_port=in_port, eth_type=ether_types.ETH_TYPE_IP, ipv4_src=src_net) # TODO: create the right match
            # verify if we have a valid buffer_id, if yes avoid to send both
            # flow_mod & packet_out
            # TODO: figure out what goes on here
            if msg.buffer_id != ofproto.OFP_NO_BUFFER:
                self.add_flow(datapath, 10, match, actions, msg.buffer_id) # TODO: make sure the priority is >1
                return
            else:
                self.add_flow(datapath, 10, match, actions)

            # TODO: Only send packet-out msg if we accept the traffic --> only if actions != []
            data = None
            if msg.buffer_id == ofproto.OFP_NO_BUFFER:
                data = msg.data

            out = parser.OFPPacketOut(datapath=datapath, buffer_id=msg.buffer_id,
                                      in_port=in_port, actions=actions, data=data)
            datapath.send_msg(out)
        

    def add_flow(self, datapath, priority, match, actions, buffer_id=None):
        ofproto = datapath.ofproto
        parser = datapath.ofproto_parser

        inst = [parser.OFPInstructionActions(ofproto.OFPIT_APPLY_ACTIONS,
                                             actions)]
        if buffer_id:
            mod = parser.OFPFlowMod(datapath=datapath, buffer_id=buffer_id,
                                    priority=priority, match=match,
                                    instructions=inst)
        else:
            mod = parser.OFPFlowMod(datapath=datapath, priority=priority,
                                    match=match, instructions=inst)
        datapath.send_msg(mod)

            
           
        

    