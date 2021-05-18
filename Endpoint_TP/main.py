from ryu.base import app_manager
from ryu.controller import ofp_event
from ryu.controller.handler import MAIN_DISPATCHER
from ryu.controller.handler import set_ev_cls
from ryu.ofproto import ofproto_v1_0, ofproto_v1_2, ofproto_v1_3, ofproto_v1_4, ofproto_v1_5

from ryu.lib.packet import packet
from ryu.lib.packet import ethernet
from ryu.lib.packet import ether_types
from ryu.lib.packet import packet
from ryu.lib.packet import ethernet
from ryu.lib.packet import ether_types
from ryu.lib.packet import in_proto
from ryu.lib.packet import ipv4
from ryu.lib.packet import icmp
from ryu.lib.packet import tcp
from ryu.lib.packet import udp

class EndpointTP(app_manager.RyuApp):
    OFP_VERSIONS = [ofproto_v1_0.OFP_VERSION]

    def __init__(self, *args, **kwargs):
        super(EndpointTP, self).__init__(*args, **kwargs)

    @set_ev_cls(ofp_event.EventOFPPacketIn, MAIN_DISPATCHER)
    def packet_in_handler(self, ev):
        
        msg = ev.msg
        datapath = msg.datapath
        ofproto = datapath.ofproto
        parser = datapath.ofproto_parser
        #in_port = msg.match['in_port']

        pkt = packet.Packet(msg.data)
        eth = pkt.get_protocols(ethernet.ethernet)[0]

        #if eth.ethertype == ether_types.ETH_TYPE_LLDP:
        #    # ignore lldp packet
        #    return
        #dst = eth.dst
        #src = eth.src

        #dpid = datapath.id
        #self.mac_to_port.setdefault(dpid, {})

        #self.logger.info("packet in %s %s %s %s", dpid, src, dst, in_port)

        ## learn a mac address to avoid FLOOD next time.
        #self.mac_to_port[dpid][src] = in_port

        #if dst in self.mac_to_port[dpid]:
        #    out_port = self.mac_to_port[dpid][dst]
        #else:
        #    out_port = ofproto.OFPP_FLOOD

        #actions = [parser.OFPActionOutput(out_port)]

        # install a flow to avoid packet_in next time
        #if out_port != ofproto.OFPP_FLOOD:

        # check IP Protocol and create a match for IP
        actions = []
        if eth.ethertype == ether_types.ETH_TYPE_IP:
            ip = pkt.get_protocol(ipv4.ipv4)
            srcip = ip.src
            dstip = ip.dst
            protocol = ip.proto
            print("######### "+str(protocol))
            print("######### "+str(srcip))
            print("######### "+str(dstip))
            print("######### "+str(pkt))
#            # if ICMP Protocol
#            if protocol == in_proto.IPPROTO_ICMP:
#                match = parser.OFPMatch(eth_type=ether_types.ETH_TYPE_IP, ipv4_src=srcip, ipv4_dst=dstip, ip_proto=protocol)
#        
#            #  if TCP Protocol
#            elif protocol == in_proto.IPPROTO_TCP:
#                t = pkt.get_protocol(tcp.tcp)
#                match = parser.OFPMatch(eth_type=ether_types.ETH_TYPE_IP, ipv4_src=srcip, ipv4_dst=dstip, ip_proto=protocol, tcp_src=t.src_port, tcp_dst=t.dst_port,)
#        
#            #  If UDP Protocol 
#            elif protocol == in_proto.IPPROTO_UDP:
#                u = pkt.get_protocol(udp.udp)
#                match = parser.OFPMatch(eth_type=ether_types.ETH_TYPE_IP, ipv4_src=srcip, ipv4_dst=dstip, ip_proto=protocol, udp_src=u.src_port, udp_dst=u.dst_port,)            
#            # verify if we have a valid buffer_id, if yes avoid to send both
#            # flow_mod & packet_out
#            if msg.buffer_id != ofproto.OFP_NO_BUFFER:
#                self.add_flow(datapath, 1, match, actions, msg.buffer_id)
#                return
#            else:
#                self.add_flow(datapath, 1, match, actions)
#        data = None
#        if msg.buffer_id == ofproto.OFP_NO_BUFFER:
#            data = msg.data
#
        #out = parser.OFPPacketOut(datapath=datapath, buffer_id=msg.buffer_id,
        #                          in_port=in_port, actions=actions, data=data)
        #datapath.send_msg(out)